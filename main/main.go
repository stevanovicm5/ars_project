package main

import (
	"alati_projekat/handlers"
	"alati_projekat/model" // Dodaj import za model za potrebe test podataka
	"alati_projekat/repository"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func setupRouter(h *handlers.ConfigHandler) *mux.Router {
	router := mux.NewRouter()

	// configuration routes
	configRouter := router.PathPrefix("/configurations").Subrouter()
	// CRUD
	configRouter.HandleFunc("", h.HandleAddConfiguration).Methods("POST")
	configRouter.HandleFunc("", h.HandleGetConfiguration).Methods("GET")
	configRouter.HandleFunc("", h.HandleUpdateConfiguration).Methods("PUT")
	configRouter.HandleFunc("", h.HandleDeleteConfiguration).Methods("DELETE")

	// configgroup routes
	groupRouter := router.PathPrefix("/configgroups").Subrouter()

	// CRUD
	groupRouter.HandleFunc("", h.HandleAddConfigurationGroup).Methods("POST")
	groupRouter.HandleFunc("", h.HandleGetConfigurationGroup).Methods("GET")
	groupRouter.HandleFunc("", h.HandleUpdateConfigurationGroup).Methods("PUT")
	groupRouter.HandleFunc("", h.HandleDeleteConfigurationGroup).Methods("DELETE")

	return router
}

func main() {
	repo := repository.NewInMemoryRepository()

	// 1. Dodajemo test podatke da bi GET radio odmah
	configV1 := model.Configuration{
		ID:      uuid.New(),
		Name:    "ServiceX", // Koristi ime iz tvog GET testa
		Version: "v1.0.0",   // Koristi verziju iz tvog GET testa
		Params:  []model.Parameter{{Key: "test", Value: "ready"}},
	}
	repo.AddConfiguration(configV1)
	// ----------------------------------------------

	configHandler := handlers.NewConfigHandler(repo)

	// ROUTER
	router := setupRouter(configHandler)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("Configuration service is running on http://localhost:8080...")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on port 8080: %v\n", err)
	}
}
