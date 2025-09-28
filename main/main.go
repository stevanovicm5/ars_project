package main

import (
	"alati_projekat/handlers"
	"alati_projekat/model"
	"alati_projekat/repository"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type application struct {
	Repo repository.Repository
}

func (app *application) IdempotencyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			next.ServeHTTP(w, r)
			return
		}

		idempotencyKey := r.Header.Get("X-Request-Id")

		if idempotencyKey == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error: X-Request-Id header (UUID) je obavezan."))
			return
		}

		isProcessed, err := app.Repo.CheckIdempotencyKey(idempotencyKey)
		if err != nil {
			log.Printf("IDEMPOTENCY ERROR: Consul check failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if isProcessed {
			log.Printf("IDEMPOTENCY HIT: Request with key %s already processed.", idempotencyKey)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("Request already processed (Idempotent)."))
			return
		}

		log.Printf("IDEMPOTENCY MISS: Processing new request with key %s.", idempotencyKey)
		next.ServeHTTP(w, r)
	})
}

func setupRouter(app *application) *mux.Router {
	router := mux.NewRouter()

	router.Use(app.IdempotencyMiddleware)

	configHandler := handlers.NewConfigHandler(app.Repo)

	// configuration routes
	configRouter := router.PathPrefix("/configurations").Subrouter()
	configRouter.HandleFunc("", configHandler.HandleAddConfiguration).Methods("POST")
	configRouter.HandleFunc("", configHandler.HandleGetConfiguration).Methods("GET")
	configRouter.HandleFunc("", configHandler.HandleUpdateConfiguration).Methods("PUT")
	configRouter.HandleFunc("", configHandler.HandleDeleteConfiguration).Methods("DELETE")

	// configgroup routes
	groupRouter := router.PathPrefix("/configgroups").Subrouter()
	groupRouter.HandleFunc("", configHandler.HandleAddConfigurationGroup).Methods("POST")
	groupRouter.HandleFunc("", configHandler.HandleGetConfigurationGroup).Methods("GET")
	groupRouter.HandleFunc("", configHandler.HandleUpdateConfigurationGroup).Methods("PUT")
	groupRouter.HandleFunc("", configHandler.HandleDeleteConfigurationGroup).Methods("DELETE")

	router.HandleFunc("/health", app.handleHealthCheck).Methods("GET")

	return router
}

func (app *application) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Configuration service is online."))
}

func main() {
	consulAddr := "http://127.0.0.1:8500"

	if os.Getenv("CONSUL_ADDR") != "" {
		consulAddr = os.Getenv("CONSUL_ADDR")
	}

	repo, err := repository.NewConsulRepository(consulAddr)

	if err != nil {
		log.Fatalf("Fatal error: Failed to connect to Consul at %s: %v", consulAddr, err)
	}
	log.Printf("Successfully connected to Consul at %s", consulAddr)

	app := &application{
		Repo: repo,
	}

	configV1 := model.Configuration{
		ID:      uuid.New(),
		Name:    "ServiceX",
		Version: "v1.0.0",
		Params:  []model.Parameter{{Key: "test", Value: "ready"}},
	}
	if err := repo.AddConfiguration(configV1); err != nil {
		log.Printf("Warning: Failed to add initial test configuration: %v", err)
	}

	router := setupRouter(app)

	port := ":8080"
	srv := &http.Server{
		Addr:         port,
		Handler:      router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// GRACEFUL SHUTDOWN
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
	}()
	// --------------------------------------------------------------------------

	log.Printf("Configuration service is running on http://localhost%s...", port)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}

	log.Println("Server exited gracefully.")
}
