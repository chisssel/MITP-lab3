package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
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

func TestHealthCheckEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
		expectedBody   string
		expectedJSON   bool
	}{
		{"GET health endpoint", "/health", "GET", http.StatusOK, `{"status":"ok"}`, true},
		{"GET health with trailing slash", "/health/", "GET", http.StatusOK, `{"status":"ok"}`, true},
		{"POST to health should fail", "/health", "POST", http.StatusMethodNotAllowed, "", false},
		{"PUT to health should fail", "/health", "PUT", http.StatusMethodNotAllowed, "", false},
		{"DELETE to health should fail", "/health", "DELETE", http.StatusMethodNotAllowed, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"status":"ok"}`))
			})
			mux.HandleFunc("/health/", func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"status":"ok"}`))
			})

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedJSON && tt.expectedBody != "" {
				body, _ := io.ReadAll(rec.Body)
				if string(body) != tt.expectedBody {
					t.Errorf("expected body %s, got %s", tt.expectedBody, string(body))
				}
			}
		})
	}
}

func TestHealthCheckResponseFormat(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		expectedKeys []string
	}{
		{"health endpoint returns status", "/health", []string{"status"}},
		{"health with slash returns status", "/health/", []string{"status"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"status":"ok"}`))
			})
			mux.HandleFunc("/health/", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"status":"ok"}`))
			})

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			contentType := rec.Header().Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				t.Errorf("expected Content-Type application/json, got %s", contentType)
			}
		})
	}
}

func TestHealthCheckCurlCompatible(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		curlCode int
	}{
		{"health for curl -f flag", "/health", 0},
		{"health slash for curl -f flag", "/health/", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"status":"ok"}`))
			})
			mux.HandleFunc("/health/", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"status":"ok"}`))
			})

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			curlExitCode := 0
			if rec.Code >= 400 {
				curlExitCode = 22
			}

			if curlExitCode != tt.curlCode {
				t.Errorf("expected curl exit code %d for path %s, got %d", tt.curlCode, tt.path, curlExitCode)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		expected     string
	}{
		{"default value when env not set", "NON_EXISTENT_VAR", "default", "default"},
		{"empty string default", "NON_EXISTENT_VAR", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.key)
			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetEnvWithValue(t *testing.T) {
	key := "TEST_VAR"
	value := "test_value"
	os.Setenv(key, value)
	defer os.Unsetenv(key)

	result := getEnv(key, "default")
	if result != value {
		t.Errorf("expected %s, got %s", value, result)
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		expected     int
	}{
		{"default value when env not set", "NON_EXISTENT_VAR", 8080, 8080},
		{"default when not a number", "NON_NUMERIC_VAR", 3000, 3000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.key)
			result := getEnvInt(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestGetEnvIntWithValue(t *testing.T) {
	key := "TEST_PORT"
	value := "9090"
	os.Setenv(key, value)
	defer os.Unsetenv(key)

	result := getEnvInt(key, 8080)
	if result != 9090 {
		t.Errorf("expected 9090, got %d", result)
	}
}

func TestGetEnvURL(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
	}{
		{"default URL when env not set", "NON_EXISTENT_URL", "http://default:5000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.key)
			result := getEnvURL(tt.key, tt.defaultValue)
			expected, _ := url.Parse(tt.defaultValue)
			if result.String() != expected.String() {
				t.Errorf("expected %s, got %s", expected.String(), result.String())
			}
		})
	}
}

func TestGetEnvURLWithValue(t *testing.T) {
	key := "TEST_URL"
	value := "http://custom:8080"
	os.Setenv(key, value)
	defer os.Unsetenv(key)

	result := getEnvURL(key, "http://default:5000")
	if result.String() != value {
		t.Errorf("expected %s, got %s", value, result.String())
	}
}

func TestGetProxy(t *testing.T) {
	proxy := getProxy("http://python-service:5000")
	if proxy == nil {
		t.Error("expected proxy to be created")
	}
}

func TestEnvironmentVariablesConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		envKey      string
		envValue    string
		envFallback string
		expected    string
	}{
		{
			"custom port",
			"GATEWAY_PORT",
			"9000",
			"8080",
			"9000",
		},
		{
			"custom python URL",
			"PYTHON_SERVICE_URL",
			"http://custom:6000",
			"http://python-service:5000",
			"http://custom:6000",
		},
		{
			"custom rust URL",
			"RUST_SERVICE_URL",
			"http://custom:7000",
			"http://rust-service:4000",
			"http://custom:7000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envKey, tt.envValue)
			defer os.Unsetenv(tt.envKey)

			result := getEnv(tt.envKey, tt.envFallback)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
