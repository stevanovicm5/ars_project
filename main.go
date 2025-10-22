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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	httpSwagger "github.com/swaggo/http-swagger"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

type application struct {
	Services services.Service
}

func initTracer() *sdktrace.TracerProvider {
	ctx := context.Background()

	otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otlpEndpoint == "" {
		otlpEndpoint = "jaeger:4317"
	}

	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(otlpEndpoint),
	)
	if err != nil {
		log.Fatalf("failed to create OTLP exporter: %v", err)
	}

	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "configuration-service-default"
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			attribute.String("environment", "docker-compose"),
		),
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp
}

// @title  Configuration API
// @version 1.0
// @description This is a configuration management API with idempotency support.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
func main() {
	tp := initTracer()
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

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
	tracingService := services.NewTracingService(baseService)
	configService := services.NewMetricsService(tracingService)

	app := &application{
		Services: configService,
	}
	configV1 := model.Configuration{
		ID:      uuid.New(),
		Name:    "ServiceX",
		Version: "v1.0.0",
		Params:  []model.Parameter{{Key: "test", Value: "ready"}},
	}

	if err := repo.AddConfiguration(context.Background(), configV1); err != nil {
		log.Printf("Warning: Failed to add initial test configuration: %v", err)
	}

	if err := repo.AddConfiguration(context.Background(), configV1); err != nil {
		log.Printf("Warning: Failed to add initial test configuration: %v", err)
	}

	router := setupRouter(app)
	rateLimiter := middleware.NewRateLimiter(4, time.Minute)
	port := ":8080"
	srv := &http.Server{
		Addr:         port,
		Handler:      rateLimiter.Middleware(router),
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

func setupRouter(app *application) *mux.Router {
	router := mux.NewRouter()

	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	router.HandleFunc("/health", app.handleHealthCheck).Methods("GET")

	configHandler := handlers.NewConfigHandler(app.Services)

	idempotencyMiddleware := middleware.NewIdempotencyMiddleware(app.Services)

	apiRouter := router.PathPrefix("/").Subrouter()

	// OVDE DODAJEMO NOVI HTTP METRICS MIDDLEWARE
	apiRouter.Use(middleware.HTTPMetricsMiddleware)

	apiRouter.Use(middleware.TracingMiddleware)
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
// @Summary Health check
// @Description Check if service is healthy
// @Tags  health
// @Accept  json
// @Produce json
// @Success 200 {object} map[string]string
// @Router  /health [get]
func (app *application) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "service": "configuration-service"}`))
}
