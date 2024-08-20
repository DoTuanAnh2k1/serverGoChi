package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"serverGoChi/models/config_models"
	"serverGoChi/src/router"
	"serverGoChi/src/server"
	"serverGoChi/src/store"
	"syscall"
)

// Server Variable
var svr *server.Server

// Init Function
func init() {
	// Set Go Log Flags
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	// Initialize Server
	svr = server.NewServer(router.Router)
}

// Main Function
func main() {
	// Init store
	store.Init(&config_models.DatabaseConfigInit)
	// Starting Server
	svr.Start()
	stopOrKillServer()
}

func stopOrKillServer() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT, os.Interrupt)
	sig := <-signals
	fmt.Println("Receive Signal from OS - Release resource")
	fmt.Println(sig)
	svr.Stop()
	os.Exit(1)
}
