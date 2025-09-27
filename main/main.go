package main

import (
	"alati_projekat/handlers"
	"alati_projekat/repository"
	"alati_projekat/model" // Dodaj import za model za potrebe test podataka
	"github.com/google/uuid" // Dodaj import za UUID za test podatke
	"fmt"
	"log"
	"net/http"
	"time"
)

// configurationsRouter je funkcija koja preusmerava zahteve na odgovarajući handler
// na osnovu HTTP metode.
func configurationsRouter(h *handlers.ConfigHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.HandleCreateConfiguration(w, r)
		case http.MethodGet:
			h.HandleGetConfiguration(w, r) // RUKUJE GET ZAHTEVIMA
		// OVI OSTALI SU POTREBNI ZA KOMPLETAN API:
		// case http.MethodDelete:
		// 	h.HandleDeleteConfiguration(w, r)
		default:
			http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		}
	}
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

	// 2. Sada registrujemo naš NOVI Router (Multiplexer),
	// koji će prosleđivati zahteve dalje.
	http.HandleFunc("/configurations", configurationsRouter(configHandler))

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("Configuration service is running on http://localhost:8080...")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on port 8080: %v\n", err)
	}
}