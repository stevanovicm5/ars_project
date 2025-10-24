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
)

// MockService for testing handlers
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

func (m *MockService) SaveIdempotencyKey(ctx context.Context, key string) {}

func (m *MockService) AddConfiguration(ctx context.Context, config model.Configuration, idempotencyKey string) error {
	key := m.makeConfigKey(config.Name, config.Version)
	if _, exists := m.configs[key]; exists {
		return errors.New("configuration already exists")
	}
	m.configs[key] = config
	return nil
}

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

func (m *MockService) DeleteConfiguration(ctx context.Context, name, version string) error {
	key := m.makeConfigKey(name, version)
	if _, exists := m.configs[key]; !exists {
		return errors.New("configuration not found")
	}
	delete(m.configs, key)
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

func (m *MockService) AddConfigurationGroup(ctx context.Context, group model.ConfigurationGroup, idempotencyKey string) error {
	key := m.makeGroupKey(group.Name, group.Version)
	if _, exists := m.groups[key]; exists {
		return errors.New("configuration group already exists")
	}
	m.groups[key] = group
	return nil
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
	return group.Configurations, nil
}

func (m *MockService) DeleteConfigsByLabels(ctx context.Context, name, version string, want map[string]string) (int, error) {
	return 0, nil
}

// Tests
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

	req := httptest.NewRequest("GET", "/configurations/get-test/v1.0.0", nil)
	rr := httptest.NewRecorder()

	handler.HandleGetConfiguration(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
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

func TestConfigHandler_GetConfiguration_MissingParams(t *testing.T) {
	mockService := NewMockService()
	handler := NewConfigHandler(mockService)

	req := httptest.NewRequest("GET", "/configurations//v1.0.0", nil)
	rr := httptest.NewRecorder()
	handler.HandleGetConfiguration(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing name, got %d", rr.Code)
	}

	req2 := httptest.NewRequest("GET", "/configurations/test/", nil)
	rr2 := httptest.NewRecorder()
	handler.HandleGetConfiguration(rr2, req2)
	if rr2.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing version, got %d", rr2.Code)
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

	req := httptest.NewRequest("PUT", "/configurations/update-test/v1.0.0", bytes.NewReader(body))
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

	req := httptest.NewRequest("DELETE", "/configurations/delete-test/v1.0.0", nil)
	rr := httptest.NewRecorder()

	handler.HandleDeleteConfiguration(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", rr.Code)
	}
}

func TestConfigHandler_WrongMethod(t *testing.T) {
	mockService := NewMockService()
	handler := NewConfigHandler(mockService)

	req := httptest.NewRequest("POST", "/configurations/test/v1.0.0", nil)
	rr := httptest.NewRecorder()

	handler.HandleGetConfiguration(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 for wrong method, got %d", rr.Code)
	}
}

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

	req := httptest.NewRequest("PUT", "/configgroups/update-group-test/v1.0.0", bytes.NewReader(body))
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
