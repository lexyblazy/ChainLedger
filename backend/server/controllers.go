package server

import (
	"encoding/json"
	"net/http"
	// "sync"
	// "polychain.capital/config"
)

func (s *Server) status(w http.ResponseWriter, r *http.Request) {
	networkStatuses, err := s.i.GetStatus(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(StatusResponse{Networks: networkStatuses})
}
