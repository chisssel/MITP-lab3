package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	pythonURL, _ := url.Parse("http://python-service:5000")
	rustURL, _ := url.Parse("http://rust-service:4000")

	pythonProxy := httputil.NewSingleHostReverseProxy(pythonURL)
	rustProxy := httputil.NewSingleHostReverseProxy(rustURL)

	originalDirector := pythonProxy.Director
	pythonProxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = "/users/"
	}

	originalDirectorRust := rustProxy.Director
	rustProxy.Director = func(req *http.Request) {
		originalDirectorRust(req)
		req.URL.Path = "/stats"
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
		w.Write([]byte(`{"status":"ok"}`))
	})

	log.Println("Go API Gateway starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
