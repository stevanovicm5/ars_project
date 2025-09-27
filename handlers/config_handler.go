package handlers

import (
	"alati_projekat/model"
	"alati_projekat/repository"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type ConfigHandler struct {
	Repo repository.ConfigRepository
}

func NewConfigHandler(repo repository.ConfigRepository) *ConfigHandler {
	return &ConfigHandler{
		Repo: repo,
	}
}

func (h *ConfigHandler) HandleCreateConfiguration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// 1. Check HTTP Method
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Decode JSON body into our Request struct
	var req model.CreateConfigurationRequest
	//r.body je reader, dekodiramo json podatke u nasu req strukturu
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body"+err.Error(), http.StatusBadRequest)
		return
	}

	//3 map request struct to full config model
	newConfig := model.Configuration{
		ID:      uuid.New(),
		Name:    req.Name,
		Version: req.Version,
		Params:  req.Params,
	}

	if err := h.Repo.AddConfiguration(newConfig); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}
func (h *ConfigHandler) HandleGetConfiguration(w http.ResponseWriter, r *http.Request) {
	// 1. Provera HTTP Metoda (iako je ruter to već uradio, ovo je dobra praksa)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Čitanje Query Parametara iz URL-a
	// Tražimo /configurations?name=ServiceA&version=v1.0.0
	name := r.URL.Query().Get("name")
	version := r.URL.Query().Get("version")

	// 3. Validacija ulaza
	if name == "" || version == "" {
		http.Error(w, "Query parameters 'name' and 'version' are required.", http.StatusBadRequest) // 400 Bad Request
		return
	}

	// 4. Pozivanje Repozitorijuma za dobavljanje podataka
	config, err := h.Repo.GetConfiguration(name, version)

	if err != nil {
		// Ako repozitorijum vrati grešku (npr. "configuration not found")
		log.Printf("Error retrieving config %s/%s: %v", name, version, err)
		http.Error(w, "Configuration not found.", http.StatusNotFound) // 404 Not Found
		return
	}

	// 5. Slanje uspešnog odgovora (200 OK)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Konvertujemo strukturu config nazad u JSON i pišemo u http.ResponseWriter
	json.NewEncoder(w).Encode(config)
}
