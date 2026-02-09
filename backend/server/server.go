package server

import (
	"context"
	"encoding/json"
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

func (s *Server) jsonHandler(fn func(r *http.Request) (any, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, err := fn(r)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(data)
	}
}

func (s *Server) SetupRoutes(router *http.ServeMux) {
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	router.HandleFunc("/status", s.jsonHandler(s.getStatus))
	router.HandleFunc("/wallets", s.jsonHandler(s.getWallets))
	router.HandleFunc("/tokens", s.jsonHandler(s.getTokens))
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
