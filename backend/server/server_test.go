package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lexyblazy/chainledger/config"
)

func TestGetOffsetPaginationParams(t *testing.T) {
	s := &Server{config: &config.Config{Api: config.ApiConfig{MaxResults: 50}}}

	t.Run("defaults", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/wallets", nil)
		limit, offset, err := s.getOffsetPaginationParams(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if limit != 50 || offset != 0 {
			t.Fatalf("got limit=%d offset=%d, want limit=50 offset=0", limit, offset)
		}
	})

	t.Run("clamps values", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/wallets?limit=200&offset=-5", nil)
		limit, offset, err := s.getOffsetPaginationParams(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if limit != 50 || offset != 0 {
			t.Fatalf("got limit=%d offset=%d, want limit=50 offset=0", limit, offset)
		}
	})

	t.Run("invalid number", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/wallets?limit=bad", nil)
		_, _, err := s.getOffsetPaginationParams(req)
		if err == nil {
			t.Fatal("expected parse error")
		}
	})
}

func TestGetCursorPaginationParams(t *testing.T) {
	s := &Server{config: &config.Config{Api: config.ApiConfig{MaxResults: 25}}}

	req := httptest.NewRequest(http.MethodGet, "/x?limit=100&cursor_id=123", nil)
	limit, cursor := s.getCursorPaginationParams(req, "cursor_id")
	if limit != 25 {
		t.Fatalf("limit=%d, want 25", limit)
	}
	if cursor != "123" {
		t.Fatalf("cursor=%v, want 123", cursor)
	}

	req = httptest.NewRequest(http.MethodGet, "/x?limit=oops", nil)
	limit, cursor = s.getCursorPaginationParams(req, "cursor_id")
	if limit != 0 || cursor != nil {
		t.Fatalf("invalid limit case got limit=%d cursor=%v, want 0,nil", limit, cursor)
	}
}

func TestCORSMiddleware(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	h := corsMiddleware(next)

	t.Run("preflight", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/status", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("status=%d, want %d", w.Code, http.StatusNoContent)
		}
		if nextCalled {
			t.Fatal("next handler should not be called for OPTIONS")
		}
		if got := w.Header().Get("Access-Control-Allow-Origin"); got == "" {
			t.Fatal("expected CORS headers to be set")
		}
	})

	t.Run("normal request", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest(http.MethodGet, "/status", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status=%d, want %d", w.Code, http.StatusOK)
		}
		if !nextCalled {
			t.Fatal("next handler should be called for non-OPTIONS requests")
		}
	})
}

func TestJSONHandler(t *testing.T) {
	s := &Server{}

	t.Run("success response", func(t *testing.T) {
		h := s.jsonHandler(func(r *http.Request) (any, int, error) {
			return map[string]bool{"ok": true}, http.StatusCreated, nil
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("status=%d, want %d", w.Code, http.StatusCreated)
		}

		var payload map[string]bool
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if !payload["ok"] {
			t.Fatalf("unexpected payload: %+v", payload)
		}
	})

	t.Run("error response", func(t *testing.T) {
		h := s.jsonHandler(func(r *http.Request) (any, int, error) {
			return nil, http.StatusBadRequest, errors.New("bad input")
		})

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status=%d, want %d", w.Code, http.StatusBadRequest)
		}

		var payload ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if payload.Error != "bad input" {
			t.Fatalf("error=%q, want bad input", payload.Error)
		}
	})
}
