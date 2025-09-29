package main

import (
	"alati_projekat/handlers"
	"alati_projekat/middleware"
	"alati_projekat/model"
	"alati_projekat/repository"
	"alati_projekat/services"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
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

	app.testRateLimiting()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}

	log.Println("Server exited gracefully.")
}

func setupRouter(app *application) *mux.Router {
	router := mux.NewRouter()

	// 1. METRICS ENDPOINT - BEZ MIDDLEWARE-A
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// 2. HEALTH CHECK - takođe bez middleware-a
	router.HandleFunc("/health", app.handleHealthCheck).Methods("GET")

	// 3. KREIRAJTE configHandler OVDE
	configHandler := handlers.NewConfigHandler(app.Services)

	// 4. KREIRAJTE POSEBNU RUTU SA MIDDLEWARE-OM ZA SVE OSTALO
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	idempotencyMiddleware := middleware.NewIdempotencyMiddleware(app.Services)

	apiRouter := router.PathPrefix("/").Subrouter()
	apiRouter.Use(rateLimiter.Middleware)
	apiRouter.Use(idempotencyMiddleware.Middleware)

	// Swagger
	staticSwaggerFiles := http.FileServer(http.Dir("./docs"))
	apiRouter.PathPrefix("/swagger/static/").Handler(http.StripPrefix("/swagger/static/", staticSwaggerFiles))
	apiRouter.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/static/swagger.json"),
	)).Methods("GET")

	// Configuration routes
	configRouter := apiRouter.PathPrefix("/configurations").Subrouter()
	configRouter.HandleFunc("", configHandler.HandleAddConfiguration).Methods("POST")
	configRouter.HandleFunc("", configHandler.HandleGetConfiguration).Methods("GET")
	configRouter.HandleFunc("", configHandler.HandleUpdateConfiguration).Methods("PUT")
	configRouter.HandleFunc("", configHandler.HandleDeleteConfiguration).Methods("DELETE")

	// Config group routes
	groupRouter := apiRouter.PathPrefix("/configgroups").Subrouter()
	groupRouter.HandleFunc("", configHandler.HandleAddConfigurationGroup).Methods("POST")
	groupRouter.HandleFunc("", configHandler.HandleGetConfigurationGroup).Methods("GET")
	groupRouter.HandleFunc("", configHandler.HandleUpdateConfigurationGroup).Methods("PUT")
	groupRouter.HandleFunc("", configHandler.HandleDeleteConfigurationGroup).Methods("DELETE")

	//Config label routes
	groupRouter.HandleFunc("/configurations", configHandler.HandleGetGroupConfigsByLabels).Methods("GET")
	groupRouter.HandleFunc("/configurations", configHandler.HandleDeleteGroupConfigsByLabels).Methods("DELETE")

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

func (app *application) testRateLimiting() {
	log.Println("Testing rate limiting...")

	testLimiter := middleware.NewRateLimiter(3, time.Minute)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	handler := testLimiter.Middleware(testHandler)

	testIP := "192.168.1.100"

	// Test 1: First 3 requests should succeed
	successCount := 0
	for i := 1; i <= 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = testIP + ":8080"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code == http.StatusOK {
			successCount++
			log.Printf("Request %d: SUCCESS (Remaining: %s)", i, rr.Header().Get("X-RateLimit-Remaining"))
		} else {
			log.Printf("Request %d: FAILED - Expected 200, got %d", i, rr.Code)
		}
	}

	// Test 2: Fourth request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = testIP + ":8080"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code == http.StatusTooManyRequests {
		log.Printf("Request 4: CORRECTLY RATE LIMITED (429)")
		log.Printf("Headers: Limit=%s, Remaining=%s, Reset=%s",
			rr.Header().Get("X-RateLimit-Limit"),
			rr.Header().Get("X-RateLimit-Remaining"),
			rr.Header().Get("X-RateLimit-Reset"))
	} else {
		log.Printf("Request 4: FAILED - Expected 429, got %d", rr.Code)
	}

	log.Println("Rate limiting test completed")
}
