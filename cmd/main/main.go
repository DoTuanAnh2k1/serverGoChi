package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"serverGoChi/config"
	"serverGoChi/models/config_models"
	"serverGoChi/src/logger"
	"serverGoChi/src/router"
	"serverGoChi/src/server"
	"serverGoChi/src/store"
	"syscall"

	"github.com/joho/godotenv"
)

// Main Function
func main() {
	// Init store
	svr := Initialize()

	// Starting Server
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
	os.Exit(1)
}

func Initialize() *server.Server {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
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
		},
		Log: config_models.LogConfig{
			Level:   os.Getenv("LOG_LEVEL"),
			DbLevel: os.Getenv("DB_LOG_LEVEL"),
		},
	}
	config.Init(cfg)
	logger.Init()

	router.Init()
	store.Init()

	// Initialize Server
	return server.NewServer(router.Router)
}
