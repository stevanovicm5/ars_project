package handlers

import (
	"alati_projekat/model"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux" // MORAMO KORISTITI MUX ZA PRAVILNU EMULACIJU VARS
)

// MockService for testing handlers
// ISPRAVLJENO: Dodate su sve metode da bi se implementirao interfejs services.Service.
type MockService struct {
	configs map[string]model.Configuration
	groups  map[string]model.ConfigurationGroup
}

func NewMockService() *MockService {
	return &MockService{
		configs: make(map[string]model.Configuration),
		groups:  make(map[string]model.ConfigurationGroup),
	}
}

func (m *MockService) makeConfigKey(name, version string) string {
	return name + ":" + version
}

func (m *MockService) makeGroupKey(name, version string) string {
	return name + ":" + version
}

func (m *MockService) CheckIdempotencyKey(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (m *MockService) SaveIdempotencyKey(ctx context.Context, key string) {
	// Nema return statement
}

func (m *MockService) AddConfiguration(ctx context.Context, config model.Configuration, idempotencyKey string) error {
	key := m.makeConfigKey(config.Name, config.Version)
	if _, exists := m.configs[key]; exists {
		return errors.New("configuration already exists")
	}
	m.configs[key] = config
	return nil
}

// ISPRAVLJENA METODA
func (m *MockService) GetConfiguration(ctx context.Context, name, version string) (model.Configuration, error) {
	key := m.makeConfigKey(name, version)
	config, exists := m.configs[key]
	if !exists {
		return model.Configuration{}, errors.New("configuration not found")
	}
	return config, nil
}

func (m *MockService) UpdateConfiguration(ctx context.Context, config model.Configuration, idempotencyKey string) (model.Configuration, error) {
	key := m.makeConfigKey(config.Name, config.Version)

	originalConfig, exists := m.configs[key]
	if !exists {
		return model.Configuration{}, errors.New("configuration not found")
	}

	if config.ID == uuid.Nil {
		config.ID = originalConfig.ID
	}

	m.configs[key] = config
	return config, nil
}

// ISPRAVLJENA METODA
func (m *MockService) DeleteConfiguration(ctx context.Context, name, version string) error {
	key := m.makeConfigKey(name, version)
	if _, exists := m.configs[key]; !exists {
		return errors.New("configuration not found")
	}
	delete(m.configs, key)
	return nil
}

func (m *MockService) AddConfigurationGroup(ctx context.Context, group model.ConfigurationGroup, idempotencyKey string) error {
	key := m.makeGroupKey(group.Name, group.Version)
	if _, exists := m.groups[key]; exists {
		return errors.New("configuration group already exists")
	}
	m.groups[key] = group
	return nil
}

func (m *MockService) GetConfigurationGroup(ctx context.Context, name, version string) (model.ConfigurationGroup, error) {
	key := m.makeGroupKey(name, version)
	group, exists := m.groups[key]
	if !exists {
		return model.ConfigurationGroup{}, errors.New("configuration group not found")
	}
	return group, nil
}

func (m *MockService) UpdateConfigurationGroup(ctx context.Context, group model.ConfigurationGroup, idempotencyKey string) (model.ConfigurationGroup, error) {
	key := m.makeGroupKey(group.Name, group.Version)
	originalGroup, exists := m.groups[key]

	if !exists {
		return model.ConfigurationGroup{}, errors.New("configuration group not found")
	}

	if group.ID == uuid.Nil {
		group.ID = originalGroup.ID
	}

	m.groups[key] = group
	return group, nil
}

func (m *MockService) DeleteConfigurationGroup(ctx context.Context, name, version string) error {
	key := m.makeGroupKey(name, version)
	if _, exists := m.groups[key]; !exists {
		return errors.New("configuration group not found")
	}
	delete(m.groups, key)
	return nil
}

func (m *MockService) FilterConfigsByLabels(ctx context.Context, name, version string, want map[string]string) ([]model.Configuration, error) {
	group, err := m.GetConfigurationGroup(ctx, name, version)
	if err != nil {
		return nil, err
	}
	// Mock implementacija za testiranje, uvek vraća sve konfiguracije
	return group.Configurations, nil
}

func (m *MockService) DeleteConfigsByLabels(ctx context.Context, name, version string, want map[string]string) (int, error) {
	// Mock implementacija, vraća 0 obrisanih
	return 0, nil
}

// -------------------------------------------------------------------
// Tests
// -------------------------------------------------------------------

func TestConfigHandler_AddConfiguration(t *testing.T) {
	mockService := NewMockService()
	handler := NewConfigHandler(mockService)

	configReq := model.CreateConfigurationRequest{
		Name:    "test-service",
		Version: "v1.0.0",
		Params:  []model.Parameter{{Key: "port", Value: "8080"}},
	}

	body, _ := json.Marshal(configReq)

	req := httptest.NewRequest("POST", "/configurations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-Id", "test-uuid-123")

	rr := httptest.NewRecorder()

	handler.HandleAddConfiguration(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rr.Code)
	}

	var response model.Configuration
	err := json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Name != "test-service" {
		t.Errorf("Expected name 'test-service', got '%s'", response.Name)
	}
}

func TestConfigHandler_AddConfiguration_BadRequest(t *testing.T) {
	mockService := NewMockService()
	handler := NewConfigHandler(mockService)

	req := httptest.NewRequest("POST", "/configurations", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.HandleAddConfiguration(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid JSON, got %d", rr.Code)
	}
}

// ISPRAVLJENI TEST ZA GET: Koristi mux.Vars
func TestConfigHandler_GetConfiguration(t *testing.T) {
	mockService := NewMockService()
	handler := NewConfigHandler(mockService)

	config := model.Configuration{
		ID:      uuid.New(),
		Name:    "get-test",
		Version: "v1.0.0",
		Params:  []model.Parameter{{Key: "test", Value: "value"}},
	}
	mockService.configs["get-test:v1.0.0"] = config

	// Kreira se HTTP zahtev za putanju /configurations/get-test/v1.0.0
	req := httptest.NewRequest("GET", "/configurations/get-test/v1.0.0", nil)

	// Dodaje Mux varijable da bi Mux.Vars() radio u handleru
	req = mux.SetURLVars(req, map[string]string{
		"name":    "get-test",
		"version": "v1.0.0",
	})

	rr := httptest.NewRecorder()
	handler.HandleGetConfiguration(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var response model.Configuration
	err := json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Name != "get-test" {
		t.Errorf("Expected name 'get-test', got '%s'", response.Name)
	}
}

// ISPRAVLJENI TEST ZA MISSING PARAMS: Koristi mux.Vars
func TestConfigHandler_GetConfiguration_MissingParams(t *testing.T) {
	mockService := NewMockService()
	handler := NewConfigHandler(mockService)

	// Test 1: Nedostaje "version" (emulira se rutingom)
	req1 := httptest.NewRequest("GET", "/configurations/test-name/", nil)
	req1 = mux.SetURLVars(req1, map[string]string{
		"name": "test-name",
		// "version" fali ili je prazan
	})
	rr1 := httptest.NewRecorder()
	handler.HandleGetConfiguration(rr1, req1)
	if rr1.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing version, got %d", rr1.Code)
	}

	// Test 2: Nedostaje "name"
	req2 := httptest.NewRequest("GET", "/configurations//v1.0.0", nil)
	req2 = mux.SetURLVars(req2, map[string]string{
		"version": "v1.0.0",
		// "name" fali ili je prazan
	})
	rr2 := httptest.NewRecorder()
	handler.HandleGetConfiguration(rr2, req2)
	if rr2.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing name, got %d", rr2.Code)
	}
}

func TestConfigHandler_UpdateConfiguration(t *testing.T) {
	mockService := NewMockService()
	handler := NewConfigHandler(mockService)
	originalID := uuid.New()

	config := model.Configuration{
		ID:      originalID,
		Name:    "update-test",
		Version: "v1.0.0",
		Params:  []model.Parameter{{Key: "old", Value: "value"}},
	}
	mockService.configs["update-test:v1.0.0"] = config

	updateReq := model.CreateConfigurationRequest{
		Name:    "update-test",
		Version: "v1.0.0",
		Params:  []model.Parameter{{Key: "new", Value: "updated-value"}},
	}

	body, _ := json.Marshal(updateReq)

	req := httptest.NewRequest("PUT", "/configurations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-Id", "update-uuid")

	rr := httptest.NewRecorder()

	handler.HandleUpdateConfiguration(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response model.Configuration
	err := json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ID != originalID {
		t.Errorf("Expected original ID %s, got %s", originalID, response.ID)
	}

	if len(response.Params) != 1 || response.Params[0].Key != "new" {
		t.Errorf("Handler did not return the updated configuration. Got params: %+v", response.Params)
	}
}

// ISPRAVLJENI TEST ZA DELETE: Koristi mux.Vars
func TestConfigHandler_DeleteConfiguration(t *testing.T) {
	mockService := NewMockService()
	handler := NewConfigHandler(mockService)

	config := model.Configuration{
		ID:      uuid.New(),
		Name:    "delete-test",
		Version: "v1.0.0",
		Params:  []model.Parameter{{Key: "test", Value: "value"}},
	}
	mockService.configs["delete-test:v1.0.0"] = config

	// Kreira se HTTP zahtev
	req := httptest.NewRequest("DELETE", "/configurations/delete-test/v1.0.0", nil)

	// Dodaje Mux varijable
	req = mux.SetURLVars(req, map[string]string{
		"name":    "delete-test",
		"version": "v1.0.0",
	})

	rr := httptest.NewRecorder()
	handler.HandleDeleteConfiguration(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("Expected status 204, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	// Provera da li je obrisano
	if _, exists := mockService.configs["delete-test:v1.0.0"]; exists {
		t.Error("Configuration was not deleted from mock service")
	}
}

func TestConfigHandler_WrongMethod(t *testing.T) {
	mockService := NewMockService()
	handler := NewConfigHandler(mockService)

	// Testiramo GET endpoint sa POST metodom
	req := httptest.NewRequest("POST", "/configurations/test/v1.0.0", nil)

	// Moramo dodati i Mux Vars za emulaciju
	req = mux.SetURLVars(req, map[string]string{
		"name":    "test",
		"version": "v1.0.0",
	})

	rr := httptest.NewRecorder()

	handler.HandleGetConfiguration(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 for wrong method, got %d", rr.Code)
	}
}

// -------------------------------------------------------------------
// Group Tests
// -------------------------------------------------------------------

func TestConfigHandler_AddConfigurationGroup(t *testing.T) {
	mockService := NewMockService()
	handler := NewConfigHandler(mockService)

	config := model.Configuration{
		ID:      uuid.New(),
		Name:    "group-config",
		Version: "v1.0.0",
		Params:  []model.Parameter{},
	}

	groupReq := model.CreateGroupRequest{
		Name:           "test-group",
		Version:        "v1.0.0",
		Configurations: []model.Configuration{config},
	}

	body, _ := json.Marshal(groupReq)

	req := httptest.NewRequest("POST", "/configgroups", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-Id", "group-uuid")

	rr := httptest.NewRecorder()

	handler.HandleAddConfigurationGroup(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rr.Code)
	}

	var response model.ConfigurationGroup
	err := json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Name != "test-group" {
		t.Errorf("Expected group name 'test-group', got '%s'", response.Name)
	}
}

func TestConfigHandler_UpdateConfigurationGroup(t *testing.T) {
	mockService := NewMockService()
	handler := NewConfigHandler(mockService)
	originalID := uuid.New()

	originalConfig := model.Configuration{ID: uuid.New(), Params: []model.Parameter{{Key: "old", Value: "val"}}}
	originalGroup := model.ConfigurationGroup{
		ID:             originalID,
		Name:           "update-group-test",
		Version:        "v1.0.0",
		Configurations: []model.Configuration{originalConfig},
	}
	mockService.groups["update-group-test:v1.0.0"] = originalGroup

	updatedConfig := model.Configuration{ID: uuid.New(), Params: []model.Parameter{{Key: "new", Value: "updated"}}}
	updateReq := model.CreateGroupRequest{
		Name:           "update-group-test",
		Version:        "v1.0.0",
		Configurations: []model.Configuration{updatedConfig},
	}

	body, _ := json.Marshal(updateReq)

	req := httptest.NewRequest("PUT", "/configgroups", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-Id", "group-update-uuid")

	rr := httptest.NewRecorder()

	handler.HandleUpdateConfigurationGroup(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var response model.ConfigurationGroup
	err := json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ID != originalID {
		t.Errorf("Expected original ID %s, got %s", originalID, response.ID)
	}

	if len(response.Configurations) != 1 || response.Configurations[0].Params[0].Key != "new" {
		t.Errorf("Handler did not return the updated group. Got configs: %+v", response.Configurations)
	}
}

// ISPRAVLJENI TEST ZA GET GROUP: Koristi mux.Vars
func TestConfigHandler_GetConfigurationGroup(t *testing.T) {
	mockService := NewMockService()
	handler := NewConfigHandler(mockService)

	group := model.ConfigurationGroup{
		ID:      uuid.New(),
		Name:    "get-group-test",
		Version: "v1.0.0",
	}
	mockService.groups["get-group-test:v1.0.0"] = group

	req := httptest.NewRequest("GET", "/configgroups/get-group-test/v1.0.0", nil)
	req = mux.SetURLVars(req, map[string]string{
		"name":    "get-group-test",
		"version": "v1.0.0",
	})
	rr := httptest.NewRecorder()

	handler.HandleGetConfigurationGroup(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var response model.ConfigurationGroup
	err := json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Name != "get-group-test" {
		t.Errorf("Expected name 'get-group-test', got '%s'", response.Name)
	}
}

// ISPRAVLJENI TEST ZA DELETE GROUP: Koristi mux.Vars
func TestConfigHandler_DeleteConfigurationGroup(t *testing.T) {
	mockService := NewMockService()
	handler := NewConfigHandler(mockService)

	group := model.ConfigurationGroup{
		ID:      uuid.New(),
		Name:    "delete-group-test",
		Version: "v1.0.0",
	}
	mockService.groups["delete-group-test:v1.0.0"] = group

	req := httptest.NewRequest("DELETE", "/configgroups/delete-group-test/v1.0.0", nil)
	req = mux.SetURLVars(req, map[string]string{
		"name":    "delete-group-test",
		"version": "v1.0.0",
	})
	rr := httptest.NewRecorder()

	handler.HandleDeleteConfigurationGroup(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("Expected status 204, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	if _, exists := mockService.groups["delete-group-test:v1.0.0"]; exists {
		t.Error("Configuration group was not deleted from mock service")
	}
}
