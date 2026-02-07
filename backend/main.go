package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"polychain.capital/config"
	"polychain.capital/db"
	"polychain.capital/ingestion"
	"polychain.capital/server"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	config, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	db, err := db.NewDB(config.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	server := server.NewServer(config.Api.Port)
	ingestion := ingestion.New(db, config)

	// Start the server in a goroutine
	go func(c context.Context) {
		server.Start(c)
	}(ctx)

	// Start the ingestion service in a goroutine
	go func(c context.Context) {
		err := ingestion.Start(c)
		if err != nil {
			log.Println("Error starting ingestion service", err)
		}
	}(ctx)

	<-ctx.Done()

	// optional shutdown with timeout
	timeout := 30 * time.Second
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), timeout)
	defer timeoutCancel()

	done := make(chan struct{})
	go func() {
		ingestion.Cleanup(timeoutCtx)
		close(done)
	}()

	select {
	case <-timeoutCtx.Done():
		log.Println("Shutdown timed out — exiting NOW")
		os.Exit(1)
	case <-done:
		log.Println("Graceful shutdown completed")
	}

}
