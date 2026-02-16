package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/lexyblazy/chainledger/config"
	"github.com/lexyblazy/chainledger/db"
	"github.com/lexyblazy/chainledger/ingestion"
)

type Server struct {
	port   string
	db     *db.DB
	config *config.Config
	i      *ingestion.StatusReader
}

func New(db *db.DB, config *config.Config, i *ingestion.StatusReader) *Server {
	return &Server{port: config.Api.Port, db: db, config: config, i: i}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) jsonHandler(fn func(r *http.Request) (any, int, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, statusCode, err := fn(r)

		w.WriteHeader(statusCode)

		if err != nil {
			var errorResponse ErrorResponse
			errorResponse.Error = err.Error()
			log.Println("Route Handler Error:", r.Method, r.URL.Path, err, "Status Code:", statusCode)
			if encodeErr := json.NewEncoder(w).Encode(errorResponse); encodeErr != nil {
				log.Println("Error writing error response:", encodeErr)
			}

			return
		}

		if encodeErr := json.NewEncoder(w).Encode(data); encodeErr != nil {
			log.Println("Error writing response:", encodeErr)
		}
	}
}

func (s *Server) SetupRoutes(router *http.ServeMux) {

	router.HandleFunc("/status", s.jsonHandler(s.getStatus))
	router.HandleFunc("/wallets", s.jsonHandler(s.handleWallets))
	router.HandleFunc("/tokens", s.jsonHandler(s.getTokens))

	router.HandleFunc("/wallets/{address}/balance-snapshots", s.jsonHandler(s.getWalletBalanceSnapshots))
	router.HandleFunc("/wallets/{address}/portfolio", s.jsonHandler(s.getWalletPortfolio))

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("Hello, World!")); err != nil {
			log.Println("Error writing root response:", err)
		}
	})

}

func (s *Server) Start(ctx context.Context) {
	log.Println("🚀 Starting server on port", s.port)
	mux := http.NewServeMux()

	s.SetupRoutes(mux)

	handler := corsMiddleware(mux)
	// TODO: add a rate limiter

	server := &http.Server{
		Addr:    ":" + s.port,
		Handler: handler,
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
