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
)

type ConfigHandler struct {
	Service services.Service
}

func NewConfigHandler(service services.Service) *ConfigHandler {
	return &ConfigHandler{
		Service: service,
	}
}

// CONFIGURATIONS

// HandleAddConfiguration godoc
//
// @Summary    Add a new configuration
// @Description  Add a new configuration with idempotency support
// @Tags      configurations
// @Accept     json
// @Produce    json
// @Param     X-Request-Id  header   string               true  "Idempotency Key (UUID)"
// @Param     request     body    model.CreateConfigurationRequest  true  "Configuration creation request"
// @Success    201       {object}  model.Configuration
// @Failure    400       {string}  string "Bad Request"
// @Failure    409       {string}  string "Conflict - Configuration already exists"
// @Failure    500       {string}  string "Internal Server Error"
// @Router     /configurations [post]
func (h *ConfigHandler) HandleAddConfiguration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.CreateConfigurationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body"+err.Error(), http.StatusBadRequest)
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

	if err := h.Service.AddConfiguration(newConfig, idempotencyKey); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newConfig)
}

// HandleGetConfiguration godoc
//
// @Summary    Get configuration
// @Description  Get configuration by name and version
// @Tags      configurations
// @Accept     json
// @Produce    json
// @Param     name  query    string true  "Configuration name"
// @Param     version query    string true  "Configuration version"
// @Success    200   {object}  model.Configuration
// @Failure    400   {string}  string "Bad Request - Name and version are required"
// @Failure    404   {string}  string "Configuration not found"
// @Failure    500   {string}  string "Internal Server Error"
// @Router     /configurations [get]
func (h *ConfigHandler) HandleGetConfiguration(w http.ResponseWriter, r *http.Request) {
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

	config, err := h.Service.GetConfiguration(name, version)

	if err != nil {
		log.Printf("Error retrieving config %s/%s: %v", name, version, err)
		http.Error(w, "Configuration not found.", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(config)
}

// HandleUpdateConfiguration godoc
//
// @Summary    Update a configuration
// @Description  Update an existing configuration with idempotency support
// @Tags      configurations
// @Accept     json
// @Produce    json
// @Param     X-Request-Id  header   string               true  "Idempotency Key (UUID)"
// @Param     request     body    model.CreateConfigurationRequest  true  "Configuration update request"
// @Success    200       {object}  model.Configuration
// @Failure    400       {string}  string "Bad Request"
// @Failure    404       {string}  string "Configuration not found"
// @Failure    500       {string}  string "Internal Server Error"
// @Router     /configurations [put]
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
		Labels:  req.Labels,
	}

	idempotencyKey := r.Header.Get("X-Request-Id")

	err := h.Service.UpdateConfiguration(updatedConfig, idempotencyKey)

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

// HandleDeleteConfiguration godoc
//
// @Summary    Delete a configuration
// @Description  Delete configuration by name and version
// @Tags      configurations
// @Accept     json
// @Produce    json
// @Param     name  query  string true  "Configuration name"
// @Param     version query  string true  "Configuration version"
// @Success    204   "No Content - Successfully deleted"
// @Failure    400   {string}  string "Bad Request - Name and version are required"
// @Failure    404   {string}  string "Configuration not found"
// @Failure    500   {string}  string "Internal Server Error"
// @Router     /configurations [delete]
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

	err := h.Service.DeleteConfiguration(name, version)

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

// HandleAddConfigurationGroup godoc
//
// @Summary    Add a new configuration group
// @Description  Add a new configuration group with idempotency support
// @Tags      configgroups
// @Accept     json
// @Produce    json
// @Param     X-Request-Id  header   string           true  "Idempotency Key (UUID)"
// @Param     request     body    model.CreateGroupRequest  true  "Configuration group creation request"
// @Success    201       {object}  model.ConfigurationGroup
// @Failure    400       {string}  string "Bad Request"
// @Failure    409       {string}  string "Conflict - Configuration group already exists"
// @Failure    500       {string}  string "Internal Server Error"
// @Router     /configgroups [post]
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

	idempotencyKey := r.Header.Get("X-Request-Id")

	if err := h.Service.AddConfigurationGroup(newGroup, idempotencyKey); err != nil {
		http.Error(w, "Group creation failed: "+err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newGroup)
}

// HandleGetConfigurationGroup godoc
//
// @Summary    Get configuration group
// @Description  Get configuration group by name and version
// @Tags      configgroups
// @Accept     json
// @Produce    json
// @Param     name  query    string true  "Configuration group name"
// @Param     version query    string true  "Configuration group version"
// @Success    200   {object}  model.ConfigurationGroup
// @Failure    400   {string}  string "Bad Request - Name and version are required"
// @Failure    404   {string}  string "Configuration group not found"
// @Failure    500   {string}  string "Internal Server Error"
// @Router     /configgroups [get]
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

	group, err := h.Service.GetConfigurationGroup(name, version)

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

// HandleUpdateConfigurationGroup godoc
//
// @Summary    Update a configuration group
// @Description  Update an existing configuration group with idempotency support
// @Tags      configgroups
// @Accept     json
// @Produce    json
// @Param     X-Request-Id  header   string           true  "Idempotency Key (UUID)"
// @Param     request     body    model.CreateGroupRequest  true  "Configuration group update request"
// @Success    200       {object}  model.ConfigurationGroup
// @Failure    400       {string}  string "Bad Request"
// @Failure    404       {string}  string "Configuration group not found"
// @Failure    500       {string}  string "Internal Server Error"
// @Router     /configgroups [put]
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

	idempotencyKey := r.Header.Get("X-Request-Id")

	err := h.Service.UpdateConfigurationGroup(updatedGroup, idempotencyKey)

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Configuration group not found for update.", http.StatusNotFound)
			return
		}
		http.Error(w, "Error updating configuration group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedGroup)
}

// HandleDeleteConfigurationGroup godoc
//
// @Summary    Delete a configuration group
// @Description  Delete configuration group by name and version
// @Tags      configgroups
// @Accept     json
// @Produce    json
// @Param     name  query  string true  "Configuration group name"
// @Param     version query  string true  "Configuration group version"
// @Success    204   "No Content - Successfully deleted"
// @Failure    400   {string}  string "Bad Request - Name and version are required"
// @Failure    404   {string}  string "Configuration group not found"
// @Failure    500   {string}  string "Internal Server Error"
// @Router     /configgroups [delete]
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

	err := h.Service.DeleteConfigurationGroup(name, version)

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

// HandleGetGroupConfigsByLabels godoc
//
// @Summary     List configurations in a group by labels
// @Description Return configurations inside a group that match ALL provided labels (format: "k:v;k2:v2")
// @Tags        configgroups
// @Accept      json
// @Produce     json
// @Param       name     query  string true  "Configuration group name"
// @Param       version  query  string true  "Configuration group version"
// @Param       labels   query  string false "Label filter: k:v;k2:v2 (all must match)"
// @Success     200 {array}  model.Configuration
// @Failure     400 {string} string "Bad Request - Name, version or labels invalid"
// @Failure     404 {string} string "Configuration group not found"
// @Failure     500 {string} string "Internal Server Error"
// @Router      /configgroups/configurations [get]
func (h *ConfigHandler) HandleGetGroupConfigsByLabels(w http.ResponseWriter, r *http.Request) {
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

	list, err := h.Service.FilterConfigsByLabels(name, version, want)
	if err != nil {
		// dosledno tvom stilu: ako sadrži "not found" → 404, inače 500
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Configuration Group not found.", http.StatusNotFound)
			return
		}
		log.Printf("Error filtering configs by labels for %s/%s: %v", name, version, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(list)
}

// HandleDeleteGroupConfigsByLabels godoc
//
// @Summary     Delete configurations in a group by labels
// @Description Delete all configurations inside a group that match ALL provided labels (format: "k:v;k2:v2")
// @Tags        configgroups
// @Accept      json
// @Produce     json
// @Param       name     query  string true  "Configuration group name"
// @Param       version  query  string true  "Configuration group version"
// @Param       labels   query  string true  "Label filter: k:v;k2:v2 (all must match)"
// @Success     200 {object} map[string]int "deleted: <count>"
// @Failure     400 {string} string "Bad Request - Name, version or labels invalid"
// @Failure     404 {string} string "Configuration group not found"
// @Failure     500 {string} string "Internal Server Error"
// @Router      /configgroups/configurations [delete]
func (h *ConfigHandler) HandleDeleteGroupConfigsByLabels(w http.ResponseWriter, r *http.Request) {
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

	deleted, err := h.Service.DeleteConfigsByLabels(name, version, want)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Configuration Group not found.", http.StatusNotFound)
			return
		}
		log.Printf("Error deleting configs by labels for %s/%s: %v", name, version, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp := map[string]int{"deleted": deleted}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
