package main

import (
	"alati_projekat/handlers"
	"alati_projekat/model"
	"alati_projekat/repository"
	"alati_projekat/services"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	httpSwagger "github.com/swaggo/http-swagger"
)

type application struct {
	Services services.Service
}

// @title   Configuration API
// @version  1.0
// @description This is a configuration management API with idempotency support.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host  localhost:8080
// @BasePath /
func main() {
	consulAddr := "http://consul:8500"

	if os.Getenv("CONSUL_HTTP_ADDR") != "" {
		consulAddr = os.Getenv("CONSUL_HTTP_ADDR")
	}
	http.Handle("/metrics", promhttp.Handler())

	repo, err := repository.NewConsulRepository(consulAddr)

	if err != nil {
		log.Fatalf("Fatal error: Failed to connect to Consul at %s: %v", consulAddr, err)
	}
	log.Printf("Successfully connected to Consul at %s", consulAddr)

	baseService := services.NewConfigurationService(repo)
	configService := services.NewMetricsService(baseService)

	app := &application{
		Services: configService,
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

	log.Printf("Configuration service is running on http://localhost%s...", port)
	log.Printf("Swagger UI available at http://localhost%s/swagger/index.html", port)
	log.Printf("Prometheus metrics available at http://localhost%s/metrics", port)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}

	log.Println("Server exited gracefully.")
}

// IdempotencyMiddleware
func (app *application) IdempotencyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			next.ServeHTTP(w, r)
			return
		}

		idempotencyKey := r.Header.Get("X-Request-Id")

		if idempotencyKey == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error: X-Request-Id header (UUID) is necessary."))
			return
		}

		isProcessed, err := app.Services.CheckIdempotencyKey(idempotencyKey)
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

// setupRouter
func setupRouter(app *application) *mux.Router {
	router := mux.NewRouter()

	router.Use(app.IdempotencyMiddleware)

	staticSwaggerFiles := http.FileServer(http.Dir("./docs"))
	router.PathPrefix("/swagger/static/").Handler(http.StripPrefix("/swagger/static/", staticSwaggerFiles))

	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/static/swagger.json"),
	)).Methods("GET")

	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	configHandler := handlers.NewConfigHandler(app.Services)

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

// HealthCheck godoc
//
// @Summary  Health check
// @Description Check if service is healthy
// @Tags   health
// @Accept   json
// @Produce  json
// @Success  200 {object} map[string]string
// @Router   /health [get]
func (app *application) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "service": "configuration-service"}`))
}
