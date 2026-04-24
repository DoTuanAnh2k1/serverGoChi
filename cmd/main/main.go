package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/DoTuanAnh2k1/serverGoChi/models/config_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/config"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/leader"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/server"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/tcpserver"
)

func main() {
	log.Println("[1/7] Loading config...")
	cfg := loadConfig()

	log.Println("[2/7] Initializing logger...")
	config.Init(cfg)
	logger.Init(cfg.Log.Level, cfg.Log.DbLevel)

	log.Println("[3/7] Connecting to database...")
	store.Init()

	log.Println("[4/7] Initializing HTTP router...")
	handler.Init()

	log.Println("[5/7] Starting TCP subscriber server...")
	startTCPServer()

	log.Println("[6/7] Starting optional services (pprof, swagger, leader)...")
	startOptionalServices(cfg)

	log.Println("[7/7] Starting HTTP server...")
	svr := server.NewServer(handler.Router)
	svr.Start()

	// Seed default data after server is running
	go func() {
		log.Println("[seed] Ensuring default user + policy exist...")
		service.SeedFirstBoot()
		log.Println("[seed] Done.")
	}()

	waitForShutdown(svr)
}

// ── Config ──────────────────────────────────────────────────────────────────

func loadConfig() *config_models.Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	expiryHours, err := strconv.Atoi(os.Getenv("JWT_EXPIRY_HOURS"))
	if err != nil || expiryHours <= 0 {
		expiryHours = 24
	}

	return &config_models.Config{
		Svr: config_models.ServerConfig{
			Host: os.Getenv("SERVER_HOST"),
			Port: os.Getenv("SERVER_PORT"),
		},
		Db: config_models.DatabaseConfig{
			DbType: os.Getenv("DB_DRIVER"),
			Mysql: config_models.MySqlConfig{
				Host:     os.Getenv("MYSQL_HOST"),
				Port:     os.Getenv("MYSQL_PORT"),
				User:     os.Getenv("MYSQL_USER"),
				Password: os.Getenv("MYSQL_PASSWORD"),
				Name:     os.Getenv("MYSQL_DB_NAME"),
			},
			Mongo: config_models.MongoConfig{
				URI:      env("MONGODB_URI", "mongodb://localhost:27017"),
				Database: env("MONGODB_DB_NAME", "cli_db"),
			},
			Postgres: config_models.PostgresConfig{
				Host:     env("POSTGRES_HOST", "localhost"),
				Port:     env("POSTGRES_PORT", "5432"),
				User:     os.Getenv("POSTGRES_USER"),
				Password: os.Getenv("POSTGRES_PASSWORD"),
				Name:     env("POSTGRES_DB_NAME", "cli_db"),
				SSLMode:  env("POSTGRES_SSLMODE", "disable"),
			},
		},
		Log: config_models.LogConfig{
			Level:   os.Getenv("LOG_LEVEL"),
			DbLevel: os.Getenv("DB_LOG_LEVEL"),
		},
		Token: config_models.TokenConfig{
			SecretKey:   os.Getenv("JWT_SECRET_KEY"),
			ExpiryHours: expiryHours,
		},
		Router: config_models.RouterConfig{
			BasePath: os.Getenv("ROUTER_BASE_PATH"),
			Origins:  env("CORS_ORIGINS", "*"),
			Methods:  env("CORS_METHODS", "GET,POST,DELETE,PUT,OPTIONS"),
			Headers:  env("CORS_HEADERS", "Content-Type,Authorization"),
		},
		Leader: loadLeaderConfig(),
	}
}

func loadLeaderConfig() config_models.LeaderConfig {
	enabled := os.Getenv("LEADER_ELECTION_ENABLED") == "true"
	leaseDuration, _ := strconv.Atoi(os.Getenv("LEASE_DURATION_SECONDS"))
	renewDeadline, _ := strconv.Atoi(os.Getenv("RENEW_DEADLINE_SECONDS"))
	retryPeriod, _ := strconv.Atoi(os.Getenv("RETRY_PERIOD_SECONDS"))
	csvExportHour, err := strconv.Atoi(os.Getenv("CSV_EXPORT_HOUR"))
	if err != nil || csvExportHour < 0 || csvExportHour > 23 {
		csvExportHour = 23
	}

	return config_models.LeaderConfig{
		Enabled:              enabled,
		LeaseName:            env("LEASE_LOCK_NAME", "mgt-service-leader"),
		Namespace:            env("LEASE_LOCK_NAMESPACE", "default"),
		PodName:              env("POD_NAME", "unknown-pod"),
		LeaseDurationSeconds: leaseDuration,
		RenewDeadlineSeconds: renewDeadline,
		RetryPeriodSeconds:   retryPeriod,
		CSVExportDir:         env("CSV_EXPORT_DIR", "/data/csv"),
		CSVExportHour:        csvExportHour,
	}
}

// ── Services ────────────────────────────────────────────────────────────────

func startTCPServer() {
	addr := ":" + env("TCP_LISTEN_PORT", "3675")
	dataDir := env("TCP_DATA_DIR", ".")
	tcp := tcpserver.New(addr, dataDir)
	if err := tcp.Start(); err != nil {
		logger.Logger.Errorf("tcp server: failed to start (continuing without it): %v", err)
		return
	}
}

func startOptionalServices(cfg *config_models.Config) {
	// pprof
	if os.Getenv("PPROF_ENABLED") == "true" {
		addr := env("PPROF_ADDR", ":6060")
		go func() {
			logger.Logger.Infof("pprof: listening on %s", addr)
			if err := http.ListenAndServe(addr, nil); err != nil {
				logger.Logger.Errorf("pprof: %v", err)
			}
		}()
	}

	// Swagger UI
	if port := os.Getenv("SWAGGER_PORT"); port != "" {
		addr := ":" + port
		go func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
				data, err := os.ReadFile(env("API_SPEC_PATH", "api.yaml"))
				if err != nil {
					http.Error(w, "api spec not found", http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/yaml")
				w.Write(data)
			})
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				fmt.Fprint(w, handler.SwaggerUIHTML("/openapi.yaml"))
			})
			logger.Logger.Infof("swagger: listening on %s", addr)
			if err := http.ListenAndServe(addr, mux); err != nil {
				logger.Logger.Errorf("swagger: %v", err)
			}
		}()
	}

	// Leader election
	if cfg.Leader.Enabled {
		go leader.Start(context.Background(), cfg.Leader, func(ctx context.Context) {
			leader.RunTasks(ctx, cfg.Leader)
		})
	} else {
		logger.Logger.Info("leader election disabled")
	}
}

// ── Shutdown ────────────────────────────────────────────────────────────────

func waitForShutdown(svr *server.Server) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)
	sig := <-signals
	log.Printf("Received %s, shutting down...", sig)
	svr.Stop()
}

// ── Helpers ─────────────────────────────────────────────────────────────────

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
