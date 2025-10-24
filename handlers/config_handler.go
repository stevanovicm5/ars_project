package handlers

import (
	"alati_projekat/labels"
	"alati_projekat/model"
	"alati_projekat/services"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type ConfigHandler struct {
	Service services.Service
}

func NewConfigHandler(service services.Service) *ConfigHandler {
	return &ConfigHandler{
		Service: service,
	}
}

var tracer = otel.Tracer("config-service-handler-tracer")

// --- CONFIGURATIONS CRUD ---

// HandleAddConfiguration godoc
// @Summary Dodaje novu konfiguraciju
// @Description Dodaje novu konfiguraciju. Koristite X-Request-Id za idempotenciju.
// @Tags configurations
// @Accept json
// @Produce json
// @Param  X-Request-Id header string false "Idempotency Key (UUID/jedinstveni ID)"
// @Param  config body model.CreateConfigurationRequest true "Telo konfiguracije"
// @Success 201 {object} model.Configuration
// @Failure 400 {string} string "Invalid request body"
// @Failure 409 {string} string "Conflict (već postoji)"
// @Router /configurations [post]
func (h *ConfigHandler) HandleAddConfiguration(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "HandleAddConfiguration")
	defer span.End()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.CreateConfigurationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.SetAttributes(attribute.String("error.message", "Invalid request body"))
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	newConfig := model.Configuration{
		ID:      uuid.New(),
		Name:    req.Name,
		Version: req.Version,
		Params:  req.Params,
		Labels:  req.Labels,
	}

	idempotencyKey := r.Header.Get("X-Request-Id")

	if err := h.Service.AddConfiguration(ctx, newConfig, idempotencyKey); err != nil {
		span.SetAttributes(attribute.String("error.message", "Conflict or Internal Error"))
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(newConfig)
}

// HandleGetConfiguration godoc
// @Summary Vraća konfiguraciju po imenu i verziji
// @Description Vraća specifičnu konfiguraciju.
// @Tags configurations
// @Produce json
// @Param name query string true "Ime konfiguracije"
// @Param version query string true "Verzija konfiguracije"
// @Success 200 {object} model.Configuration
// @Failure 400 {string} string "Missing query parameters"
// @Failure 404 {string} string "Configuration not found"
// @Router /configurations [get]
func (h *ConfigHandler) HandleGetConfiguration(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "HandleGetConfiguration")
	defer span.End()

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

	config, err := h.Service.GetConfiguration(ctx, name, version)
	if err != nil {
		log.Printf("Error retrieving config %s/%s: %v", name, version, err)
		http.Error(w, "Configuration not found.", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(config)
}

// HandleUpdateConfiguration godoc
// @Summary Ažurira postojeću konfiguraciju
// @Description Ažurira konfiguraciju. Koristite X-Request-Id za idempotenciju.
// @Tags configurations
// @Accept json
// @Produce json
// @Param  X-Request-Id header string false "Idempotency Key (UUID/jedinstveni ID)"
// @Param  config body model.CreateConfigurationRequest true "Ažurirano telo konfiguracije"
// @Success 200 {object} model.Configuration
// @Failure 400 {string} string "Invalid request body"
// @Failure 404 {string} string "Configuration not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /configurations [put]
func (h *ConfigHandler) HandleUpdateConfiguration(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "HandleUpdateConfiguration")
	defer span.End()

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.CreateConfigurationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	configToUpdate := model.Configuration{
		Name:    req.Name,
		Version: req.Version,
		Params:  req.Params,
		Labels:  req.Labels,
	}

	idempotencyKey := r.Header.Get("X-Request-Id")
	finalConfig, err := h.Service.UpdateConfiguration(ctx, configToUpdate, idempotencyKey)

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Configuration not found for update.", http.StatusNotFound)
			return
		}
		http.Error(w, "Error updating configuration: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(finalConfig)
}

// HandleDeleteConfiguration godoc
// @Summary Briše konfiguraciju
// @Description Briše specifičnu konfiguraciju po imenu i verziji.
// @Tags configurations
// @Param name query string true "Ime konfiguracije"
// @Param version query string true "Verzija konfiguracije"
// @Success 204 "No Content"
// @Failure 400 {string} string "Missing query parameters"
// @Failure 404 {string} string "Configuration not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /configurations [delete]
func (h *ConfigHandler) HandleDeleteConfiguration(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "HandleDeleteConfiguration")
	defer span.End()

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

	err := h.Service.DeleteConfiguration(ctx, name, version)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Configuration not found for deletion.", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error during deletion.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- CONFIGURATION GROUPS CRUD ---

// HandleAddConfigurationGroup godoc
// @Summary Dodaje novu grupu konfiguracija
// @Description Dodaje novu grupu konfiguracija. Koristite X-Request-Id za idempotenciju.
// @Tags configuration_groups
// @Accept json
// @Produce json
// @Param  X-Request-Id header string false "Idempotency Key (UUID/jedinstveni ID)"
// @Param  group body model.CreateGroupRequest true "Telo grupe konfiguracija"
// @Success 201 {object} model.ConfigurationGroup
// @Failure 400 {string} string "Invalid request body"
// @Failure 409 {string} string "Group creation failed (Conflict)"
// @Router /configgroups [post]
func (h *ConfigHandler) HandleAddConfigurationGroup(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "HandleAddConfigurationGroup")
	defer span.End()

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

	idempotencyKey := r.Header.Get("X-Request-Id")
	if err := h.Service.AddConfigurationGroup(ctx, newGroup, idempotencyKey); err != nil {
		http.Error(w, "Group creation failed: "+err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(newGroup)
}

// HandleGetConfigurationGroup godoc
// @Summary Vraća grupu konfiguracija po imenu i verziji
// @Description Vraća specifičnu grupu konfiguracija.
// @Tags configuration_groups
// @Produce json
// @Param name query string true "Ime grupe"
// @Param version query string true "Verzija grupe"
// @Success 200 {object} model.ConfigurationGroup
// @Failure 400 {string} string "Missing query parameters"
// @Failure 404 {string} string "Configuration group not found"
// @Router /configgroups [get]
func (h *ConfigHandler) HandleGetConfigurationGroup(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "HandleGetConfigurationGroup")
	defer span.End()

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

	group, err := h.Service.GetConfigurationGroup(ctx, name, version)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Configuration group not found.", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(group)
}

// HandleUpdateConfigurationGroup godoc
// @Summary Ažurira postojeću grupu konfiguracija
// @Description Ažurira grupu konfiguracija. Koristite X-Request-Id za idempotenciju.
// @Tags configuration_groups
// @Accept json
// @Produce json
// @Param X-Request-Id header string false "Idempotency Key (UUID/jedinstveni ID)"
// @Param group body model.CreateGroupRequest true "Ažurirano telo grupe konfiguracija"
// @Success 200 {object} model.ConfigurationGroup
// @Failure 400 {string} string "Invalid request body"
// @Failure 404 {string} string "Configuration group not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /configgroups [put]
func (h *ConfigHandler) HandleUpdateConfigurationGroup(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "HandleUpdateConfigurationGroup")
	defer span.End()

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	groupToUpdate := model.ConfigurationGroup{
		Name:           req.Name,
		Version:        req.Version,
		Configurations: req.Configurations,
	}

	idempotencyKey := r.Header.Get("X-Request-Id")
	finalGroup, err := h.Service.UpdateConfigurationGroup(ctx, groupToUpdate, idempotencyKey)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Configuration group not found for update.", http.StatusNotFound)
			return
		}
		http.Error(w, "Error updating configuration group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(finalGroup)
}

// HandleDeleteConfigurationGroup godoc
// @Summary Briše grupu konfiguracija
// @Description Briše specifičnu grupu konfiguracija po imenu i verziji.
// @Tags configuration_groups
// @Param name query string true "Ime grupe"
// @Param version query string true "Verzija grupe"
// @Success 204 "No Content"
// @Failure 400 {string} string "Missing query parameters"
// @Failure 404 {string} string "Configuration group not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /configgroups [delete]
func (h *ConfigHandler) HandleDeleteConfigurationGroup(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "HandleDeleteConfigurationGroup")
	defer span.End()

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

	err := h.Service.DeleteConfigurationGroup(ctx, name, version)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Configuration group not found for deletion.", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error during deletion.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- LABEL-BASED OPERATIONS ---

// HandleGetGroupConfigsByLabels godoc
// @Summary Filtrira konfiguracije unutar grupe po labelama
// @Description Vraća listu konfiguracija unutar grupe koje odgovaraju zadatim labelama.
// @Tags configuration_groups
// @Produce json
// @Param name query string true "Ime grupe"
// @Param version query string true "Verzija grupe"
// @Param labels query string true "Labeli (k:v;k2:v2)"
// @Success 200 {array} model.Configuration "Filtrirana lista konfiguracija"
// @Failure 400 {string} string "Missing query parameters or invalid labels format"
// @Failure 404 {string} string "Configuration Group not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /configgroups/configurations [get]
func (h *ConfigHandler) HandleGetGroupConfigsByLabels(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "HandleGetGroupConfigsByLabels")
	defer span.End()

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	version := r.URL.Query().Get("version")
	labelsRaw := r.URL.Query().Get("labels")

	if name == "" || version == "" {
		http.Error(w, "Query parameters 'name' and 'version' are required.", http.StatusBadRequest)
		return
	}

	want, err := labels.Parse(labelsRaw)
	if err != nil {
		http.Error(w, "Invalid 'labels' query: "+err.Error(), http.StatusBadRequest)
		return
	}

	list, err := h.Service.FilterConfigsByLabels(ctx, name, version, want)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Configuration Group not found.", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

// HandleDeleteGroupConfigsByLabels godoc
// @Summary Briše konfiguracije unutar grupe po labelama
// @Description Briše konfiguracije unutar grupe koje odgovaraju zadatim labelama. Vraća broj obrisanih.
// @Tags configuration_groups
// @Produce json
// @Param name query string true "Ime grupe"
// @Param version query string true "Verzija grupe"
// @Param labels query string true "Labeli (k:v;k2:v2)"
// @Success 200 {object} object{deleted=int}
// @Failure 400 {string} string "Missing query parameters or invalid labels format"
// @Failure 404 {string} string "Configuration Group not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /configgroups/configurations [delete]
func (h *ConfigHandler) HandleDeleteGroupConfigsByLabels(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "HandleDeleteGroupConfigsByLabels")
	defer span.End()

	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	version := r.URL.Query().Get("version")
	labelsRaw := r.URL.Query().Get("labels")

	if name == "" || version == "" {
		http.Error(w, "Query parameters 'name' and 'version' are required.", http.StatusBadRequest)
		return
	}
	if labelsRaw == "" {
		http.Error(w, "Query parameter 'labels' is required (format: k:v;k2:v2).", http.StatusBadRequest)
		return
	}

	want, err := labels.Parse(labelsRaw)
	if err != nil || len(want) == 0 {
		http.Error(w, "Invalid 'labels' query: expected k:v;k2:v2", http.StatusBadRequest)
		return
	}

	deleted, err := h.Service.DeleteConfigsByLabels(ctx, name, version, want)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Configuration Group not found.", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp := map[string]int{"deleted": deleted}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
