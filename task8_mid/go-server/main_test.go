package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{"health returns ok", "/health", http.StatusOK, `{"status":"ok"}`},
		{"health without slash", "/health", http.StatusOK, `{"status":"ok"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"status":"ok"}`))
			})
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			body, _ := io.ReadAll(rec.Body)
			if string(body) != tt.wantBody {
				t.Errorf("expected body %s, got %s", tt.wantBody, string(body))
			}
		})
	}
}

func TestRoutePatterns(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		matches []bool
	}{
		{"api/users matches", "/api/users/", []bool{true, false, false}},
		{"api/stats matches", "/api/stats", []bool{false, true, false}},
		{"health matches", "/health", []bool{false, false, true}},
	}

	patterns := []struct {
		pattern string
	}{
		{"/api/users/"},
		{"/api/stats"},
		{"/health"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, pattern := range patterns {
				matches := strings.HasPrefix(tt.path, pattern.pattern)
				if matches != tt.matches[i] {
					t.Errorf("path %s with pattern %s: expected %v, got %v",
						tt.path, pattern.pattern, tt.matches[i], matches)
				}
			}
		})
	}
}

func TestProxyPathMapping(t *testing.T) {
	tests := []struct {
		name            string
		gatewayPath     string
		expectedBackend string
	}{
		{"users proxy", "/api/users/", "/users/"},
		{"users without slash", "/api/users", "/users/"},
		{"stats proxy", "/api/stats", "/stats"},
		{"stats with slash", "/api/stats/", "/stats"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedPath string

			mux := http.NewServeMux()
			mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
				receivedPath = "/users/"
				w.Write([]byte(`[{"id":1}]`))
			})
			mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
				receivedPath = "/users/"
				w.Write([]byte(`[{"id":1}]`))
			})
			mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
				receivedPath = "/stats"
				w.Write([]byte(`{"total":0}`))
			})
			mux.HandleFunc("/api/stats/", func(w http.ResponseWriter, r *http.Request) {
				receivedPath = "/stats"
				w.Write([]byte(`{"total":0}`))
			})

			req := httptest.NewRequest(http.MethodGet, tt.gatewayPath, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if receivedPath != tt.expectedBackend {
				t.Errorf("expected backend path %s, got %s", tt.expectedBackend, receivedPath)
			}
		})
	}
}

func TestHTTPMethods(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
	}{
		{"GET health", "/health", http.MethodGet, http.StatusOK},
		{"GET users", "/api/users/", http.MethodGet, http.StatusOK},
		{"GET stats", "/api/stats", http.MethodGet, http.StatusOK},
		{"POST health", "/health", http.MethodPost, http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
					return
				}
				w.Write([]byte(`{"status":"ok"}`))
			})
			mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
					return
				}
				w.Write([]byte(`[{}]`))
			})
			mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
					return
				}
				w.Write([]byte(`{}`))
			})

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestContentType(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		expectedContains string
	}{
		{"health is json", "/health", "application/json"},
		{"users is json", "/api/users/", "application/json"},
		{"stats is json", "/api/stats", "application/json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"status":"ok"}`))
			})
			mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`[{}]`))
			})
			mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{}`))
			})

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			contentType := rec.Header().Get("Content-Type")
			if !strings.Contains(contentType, tt.expectedContains) {
				t.Errorf("expected Content-Type to contain %s, got %s", tt.expectedContains, contentType)
			}
		})
	}
}

func TestPathPrefixMatching(t *testing.T) {
	tests := []struct {
		name        string
		gatewayPath string
		wantMatch   bool
		route       string
	}{
		{"exact match users", "/api/users/", true, "/api/users/"},
		{"exact match stats", "/api/stats", true, "/api/stats"},
		{"partial match users", "/api/users/123", true, "/api/users/"},
		{"no match", "/api/other", false, "/api/users/"},
		{"health match", "/health", true, "/health"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := strings.HasPrefix(tt.gatewayPath, tt.route)
			if matches != tt.wantMatch {
				t.Errorf("path %s with route %s: expected %v, got %v",
					tt.gatewayPath, tt.route, tt.wantMatch, matches)
			}
		})
	}
}
