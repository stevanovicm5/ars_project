package handlers

import (
	"alati_projekat/model"
	"alati_projekat/repository"
	"encoding/json"
	"log"
	"net/http"
	"strings"

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

// CONFIGURATIONS

// ADD
func (h *ConfigHandler) HandleAddConfiguration(w http.ResponseWriter, r *http.Request) {
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

// GET
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

// UPDATE
func (h *ConfigHandler) HandleUpdateConfiguration(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.CreateConfigurationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	updatedConfig := model.Configuration{
		ID:      uuid.New(),
		Name:    req.Name,
		Version: req.Version,
		Params:  req.Params,
	}

	err := h.Repo.UpdateConfiguration(updatedConfig)

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Configuration not found for update.", http.StatusNotFound)
			return
		}
		http.Error(w, "Error updating configuration: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedConfig)
}

// DELETE
func (h *ConfigHandler) HandleDeleteConfiguration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	version := r.URL.Query().Get("version")

	if name == "" || version == "" {
		http.Error(w, "Query parameters 'name' and 'version' are required.", http.StatusBadRequest)
		return
	}

	err := h.Repo.DeleteConfiguration(name, version)

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Configuration not found for deletion.", http.StatusNotFound)
			return
		}
		log.Printf("Error deleting config %s/%s: %v", name, version, err)
		http.Error(w, "Internal server error during deletion.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CONFIGURATION GROUPS

// ADD
func (h *ConfigHandler) HandleAddConfigurationGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	newGroup := model.ConfigurationGroup{
		ID:             uuid.New(),
		Name:           req.Name,
		Version:        req.Version,
		Configurations: req.Configurations,
	}

	if err := h.Repo.AddConfigurationGroup(newGroup); err != nil {
		http.Error(w, "Group creation failed: "+err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newGroup)
}

// GET
func (h *ConfigHandler) HandleGetConfigurationGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	version := r.URL.Query().Get("version")

	if name == "" || version == "" {
		http.Error(w, "Query parameters 'name' and 'version' are required.", http.StatusBadRequest)
		return
	}

	group, err := h.Repo.GetConfigurationGroup(name, version)

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("Error retrieving group %s/%s: %v", name, version, err)
			http.Error(w, "Configuration Group not found.", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(group)
}

// UPDATE
func (h *ConfigHandler) HandleUpdateConfigurationGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	updatedGroup := model.ConfigurationGroup{
		ID:             uuid.New(),
		Name:           req.Name,
		Version:        req.Version,
		Configurations: req.Configurations,
	}

	err := h.Repo.UpdateConfigurationGroup(updatedGroup)

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Configuration group not found for update.", http.StatusNotFound)
			return
		}
		http.Error(w, "Error updating configuration group: "+err.Error(), http.StatusInternalServerError) // 500
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedGroup)
}

// DELETE
func (h *ConfigHandler) HandleDeleteConfigurationGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	version := r.URL.Query().Get("version")

	if name == "" || version == "" {
		http.Error(w, "Query parameters 'name' and 'version' are required.", http.StatusBadRequest)
		return
	}

	err := h.Repo.DeleteConfigurationGroup(name, version)

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Configuration group not found for deletion.", http.StatusNotFound)
			return
		}
		log.Printf("Error deleting group %s/%s: %v", name, version, err)
		http.Error(w, "Internal server error during deletion.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
