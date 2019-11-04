package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func unrestrictedMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func (s *server) unrestrictedAPI() {
	unrestrictedRequests := s.router.PathPrefix("/api").Subrouter()
	unrestrictedRequests.Use(unrestrictedMiddleware)

	// REGISTRATION
	unrestrictedRequests.HandleFunc("/registration", s.handleRegistration()).Methods("POST", "OPTIONS")
	// LOGIN
	unrestrictedRequests.HandleFunc("/login", s.handleLogin()).Methods("POST", "OPTIONS")
	// CHECK EMAIL EXISTS
	unrestrictedRequests.HandleFunc("/user/exists/{email}", s.handleUserExists()).Methods("GET", "OPTIONS")
}

// Connect to CockroachDB
func (s *server) ConnectCRDB() {
	godotenv.Load()

	var (
		port      = intEnv("PORT", 8000)
		originStr = env("ORIGIN", fmt.Sprintf("http://localhost:%d", port))
		dbURL     = env("DATABASE_URL", "postgresql://root@localhost:26257/?sslmode=disable")
	)

	flag.IntVar(&port, "p", port, "Port ($PORT)")
	flag.StringVar(&originStr, "origin", originStr, "Origin ($ORIGIN)")
	flag.StringVar(&dbURL, "db", dbURL, "Database URL ($DATABASE_URL)")
	flag.Parse()

	var err error
	if s.origin, err = url.Parse(originStr); err != nil || !s.origin.IsAbs() {
		log.Fatalln("invalid origin")
		fmt.Printf(err) // TODO: Create Logging for HTTP tracing
		return
	}

	if i, err := strconv.Atoi(s.origin.Port()); err == nil {
		port = i
	}

	if s.db, err = sql.Open("postgres", dbURL); err != nil {
		fmt.Println(err)
		fmt.Printf("could not open database connection: %v\n", err)
		return
	} else {
		fmt.Printf("opened database connection for: %v\n", s.db)
	}
}

// Start start api service
func (s *server) Start() {
	// handler := cors.Default().Handler(s.router)
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost", "http://localhost:3000", "http://localhost:5000", MustEnv("CORS_ENV_HTTP_URL"), MustEnv("CORS_ENV_HTTPS_URL")},
		AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS", "DELETE"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            true,
	})
	handler := cors.Handler(s.router)

	s.unrestrictedAPI()

	fmt.Println("accepting http connections on port: ", ":8000")
	log.Fatal(http.ListenAndServe(":8000", handler))
}

