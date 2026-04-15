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
	svr := Initialize()
	svr.Start()
	stopOrKillServer(svr)
}

func stopOrKillServer(svr *server.Server) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT, os.Interrupt)
	sig := <-signals
	fmt.Println("Receive Signal from OS - Release resource")
	fmt.Println(sig)
	svr.Stop()
	os.Exit(0)
}

func Initialize() *server.Server {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	expiryHours, err := strconv.Atoi(os.Getenv("JWT_EXPIRY_HOURS"))
	if err != nil || expiryHours <= 0 {
		expiryHours = 24
	}

	cfg := &config_models.Config{
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
				URI:      getEnvOrDefault("MONGODB_URI", "mongodb://localhost:27017"),
				Database: getEnvOrDefault("MONGODB_DB_NAME", "cli_db"),
			},
			Postgres: config_models.PostgresConfig{
				Host:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
				Port:     getEnvOrDefault("POSTGRES_PORT", "5432"),
				User:     os.Getenv("POSTGRES_USER"),
				Password: os.Getenv("POSTGRES_PASSWORD"),
				Name:     getEnvOrDefault("POSTGRES_DB_NAME", "cli_db"),
				SSLMode:  getEnvOrDefault("POSTGRES_SSLMODE", "disable"),
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
			Origins:  getEnvOrDefault("CORS_ORIGINS", "*"),
			Methods:  getEnvOrDefault("CORS_METHODS", "GET,POST,DELETE,PUT,OPTIONS"),
			Headers:  getEnvOrDefault("CORS_HEADERS", "Content-Type,Authorization"),
		},
		Leader: buildLeaderConfig(),
	}

	config.Init(cfg)
	logger.Init(cfg.Log.Level, cfg.Log.DbLevel)
	handler.Init()
	store.Init()
	service.SeedDefaultUser()
	service.SeedDefaultNes()

	tcpAddr := ":" + getEnvOrDefault("TCP_LISTEN_PORT", "3675")
	tcpDataDir := getEnvOrDefault("TCP_DATA_DIR", ".")
	tcp := tcpserver.New(tcpAddr, tcpDataDir)
	if err := tcp.Start(); err != nil {
		log.Fatalf("tcp server: %v", err)
	}

	// pprof server on port 6060 (only if PPROF_ENABLED=true)
	if os.Getenv("PPROF_ENABLED") == "true" {
		pprofAddr := getEnvOrDefault("PPROF_ADDR", ":6060")
		go func() {
			logger.Logger.Infof("pprof: listening on %s", pprofAddr)
			if err := http.ListenAndServe(pprofAddr, nil); err != nil {
				logger.Logger.Errorf("pprof: %v", err)
			}
		}()
	}

	// Swagger UI server (only if SWAGGER_PORT is set)
	if swaggerPort := os.Getenv("SWAGGER_PORT"); swaggerPort != "" {
		swaggerAddr := ":" + swaggerPort
		go func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
				specPath := getEnvOrDefault("API_SPEC_PATH", "api.yaml")
				data, err := os.ReadFile(specPath)
				if err != nil {
					http.Error(w, "api spec not found", http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/yaml")
				w.WriteHeader(http.StatusOK)
				w.Write(data)
			})
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				fmt.Fprint(w, handler.SwaggerUIHTML("/openapi.yaml"))
			})
			logger.Logger.Infof("swagger: listening on %s", swaggerAddr)
			if err := http.ListenAndServe(swaggerAddr, mux); err != nil {
				logger.Logger.Errorf("swagger: %v", err)
			}
		}()
	}

	if cfg.Leader.Enabled {
		ctx := context.Background()
		go leader.Start(ctx, cfg.Leader, func(leaderCtx context.Context) {
			leader.RunTasks(leaderCtx, cfg.Leader)
		})
	} else {
		logger.Logger.Info("leader election disabled (LEADER_ELECTION_ENABLED=false)")
	}

	return server.NewServer(handler.Router)
}

func buildLeaderConfig() config_models.LeaderConfig {
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
		LeaseName:            getEnvOrDefault("LEASE_LOCK_NAME", "mgt-service-leader"),
		Namespace:            getEnvOrDefault("LEASE_LOCK_NAMESPACE", "default"),
		PodName:              getEnvOrDefault("POD_NAME", "unknown-pod"),
		LeaseDurationSeconds: leaseDuration,
		RenewDeadlineSeconds: renewDeadline,
		RetryPeriodSeconds:   retryPeriod,
		CSVExportDir:         getEnvOrDefault("CSV_EXPORT_DIR", "/data/csv"),
		CSVExportHour:        csvExportHour,
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
