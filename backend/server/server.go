package server

import (
	"context"
	"log"
	"net/http"

	"polychain.capital/config"
	"polychain.capital/db"
	"polychain.capital/ingestion"
)

type Server struct {
	port   string
	db     *db.DB
	config *config.Config
	i      *ingestion.IngestionService
}

func New(db *db.DB, config *config.Config, i *ingestion.IngestionService) *Server {
	return &Server{port: config.Api.Port, db: db, config: config, i: i}
}

func (s *Server) SetupRoutes(router *http.ServeMux) {
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	router.HandleFunc("/status", s.status)
}

func (s *Server) Start(ctx context.Context) {
	log.Println("Starting server on port", s.port)
	router := http.NewServeMux()

	s.SetupRoutes(router)

	server := &http.Server{
		Addr:    ":" + s.port,
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	if err := server.Shutdown(ctx); err != nil {
		log.Println("Error shutting down server:", err)
	}
}
