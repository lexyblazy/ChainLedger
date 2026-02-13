package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lexyblazy/chainledger/config"
	"github.com/lexyblazy/chainledger/db"
	"github.com/lexyblazy/chainledger/ingestion"
	"github.com/lexyblazy/chainledger/server"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatal("mode required: api | worker")
	}
	mode := os.Args[1]
	if mode != "api" && mode != "worker" {
		log.Fatal("Invalid mode")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	config, err := config.LoadConfig("./config.json")
	if err != nil {
		log.Fatal(err)
	}

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		dbUrl = config.DatabaseURL
		log.Println("DATABASE_URL is not set, using default from config.json")
	}

	db, err := db.NewDB(dbUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	switch mode {
	case "api":
		runApi(ctx, db, config)
	case "worker":
		runWorker(ctx, db, config)
	default:
		log.Fatal("Invalid mode", mode)
	}
}

func runApi(ctx context.Context, db *db.DB, config *config.Config) {
	statusReader := ingestion.NewStatusReader(db, config)
	server := server.New(db, config, statusReader)

	// Start the server in a goroutine
	go func(c context.Context) {
		server.Start(c)
	}(ctx)

	<-ctx.Done()

	shutDownWithTimeout(ctx, time.Duration(config.ShutdownTimeoutSeconds)*time.Second, func(ctx context.Context) {
		// simulate some cleanup work here
		time.Sleep(1 * time.Second)
	})

}

func runWorker(ctx context.Context, db *db.DB, config *config.Config) {

	engine := ingestion.New(db, config)

	// Start the ingestion service in a goroutine
	go func(c context.Context) {
		err := engine.Start(c)
		if err != nil {
			log.Println("Error starting ingestion service", err)
		}
	}(ctx)

	<-ctx.Done()

	shutDownWithTimeout(ctx, time.Duration(config.ShutdownTimeoutSeconds)*time.Second, engine.Cleanup)
}

func shutDownWithTimeout(ctx context.Context, timeout time.Duration, cleanup func(ctx context.Context)) {
	// optional shutdown with timeout
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), timeout)
	defer timeoutCancel()

	done := make(chan struct{})
	go func() {
		cleanup(timeoutCtx)
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
