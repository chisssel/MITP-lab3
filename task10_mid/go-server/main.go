package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvURL(key, defaultValue string) *url.URL {
	parsed, err := url.Parse(getEnv(key, defaultValue))
	if err != nil {
		log.Printf("Failed to parse URL from %s: %v, using default", key, err)
		parsed, _ = url.Parse(defaultValue)
	}
	return parsed
}

func getProxy(target string) *httputil.ReverseProxy {
	targetURL, _ := url.Parse(target)
	return httputil.NewSingleHostReverseProxy(targetURL)
}

func main() {
	gatewayPort := getEnv("GATEWAY_PORT", "8080")
	pythonServiceURL := getEnv("PYTHON_SERVICE_URL", "http://python-service:5000")
	rustServiceURL := getEnv("RUST_SERVICE_URL", "http://rust-service:4000")
	usersPath := getEnv("USERS_PATH", "/users/")
	statsPath := getEnv("STATS_PATH", "/stats")

	log.Printf("Configuration:")
	log.Printf("  Gateway Port: %s", gatewayPort)
	log.Printf("  Python Service URL: %s", pythonServiceURL)
	log.Printf("  Rust Service URL: %s", rustServiceURL)
	log.Printf("  Users Path: %s", usersPath)
	log.Printf("  Stats Path: %s", statsPath)

	pythonProxy := getProxy(pythonServiceURL)
	rustProxy := getProxy(rustServiceURL)

	originalPythonDirector := pythonProxy.Director
	pythonProxy.Director = func(req *http.Request) {
		originalPythonDirector(req)
		req.URL.Path = usersPath
	}

	originalRustDirector := rustProxy.Director
	rustProxy.Director = func(req *http.Request) {
		originalRustDirector(req)
		req.URL.Path = statsPath
	}

	http.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		pythonProxy.ServeHTTP(w, r)
	})

	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		rustProxy.ServeHTTP(w, r)
	})

	http.HandleFunc("/api/stats/", func(w http.ResponseWriter, r *http.Request) {
		rustProxy.ServeHTTP(w, r)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"go-gateway"}`))
	})

	http.HandleFunc("/health/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"go-gateway"}`))
	})

	log.Printf("Go API Gateway starting on :%s", gatewayPort)
	addr := ":" + gatewayPort
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
