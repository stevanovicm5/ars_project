package main

import (
	"alati_projekat/handlers"
	"alati_projekat/model"
	"alati_projekat/repository"
	"log"
	"net/http"

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
	consulAddr := "127.0.0.1:8500"

	// Consul Repo
	repo, err := repository.NewConsulRepository(consulAddr)

	// ///////////////////////////////////////////////////////////////////////////////////
	// // inMemoryRepo
	// // repo := repository.NewInMemoryRepository()
	// ///////////////////////////////////////////////////////////////////////////////////

	if err != nil {
		log.Fatalf("Fatal error: Failed to connect to Consul at %s: %v", consulAddr, err)
	}
	log.Printf("Successfully connected to Consul at %s", consulAddr)

	configV1 := model.Configuration{
		ID:      uuid.New(),
		Name:    "ServiceX",
		Version: "v1.0.0",
		Params:  []model.Parameter{{Key: "test", Value: "ready"}},
	}

	if err := repo.AddConfiguration(configV1); err != nil {
		log.Printf("Warning: Failed to add initial test configuration: %v", err)
	}

	configHandler := handlers.NewConfigHandler(repo)
	router := setupRouter(configHandler)

	log.Println("Configuration service is running on http://localhost:8080...")

	log.Fatal(http.ListenAndServe(":8080", router))
}
