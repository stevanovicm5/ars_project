package services

import (
	"alati_projekat/model"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// MockRepository for testing
type MockRepository struct {
	configs         map[string]model.Configuration
	groups          map[string]model.ConfigurationGroup
	idempotencyKeys map[string]bool
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		configs:         make(map[string]model.Configuration),
		groups:          make(map[string]model.ConfigurationGroup),
		idempotencyKeys: make(map[string]bool),
	}
}

func (m *MockRepository) makeConfigKey(name, version string) string {
	return name + ":" + version
}

func (m *MockRepository) makeGroupKey(name, version string) string {
	return name + ":" + version
}

// Repository interface implementation
func (m *MockRepository) CheckIdempotencyKey(ctx context.Context, key string) (bool, error) {
	return m.idempotencyKeys[key], nil
}

func (m *MockRepository) SaveIdempotencyKey(ctx context.Context, key string) error {
	m.idempotencyKeys[key] = true
	return nil
}

func (m *MockRepository) AddConfiguration(ctx context.Context, config model.Configuration) error {
	key := m.makeConfigKey(config.Name, config.Version)
	if _, exists := m.configs[key]; exists {
		return errors.New("configuration already exists")
	}
	m.configs[key] = config
	return nil
}

func (m *MockRepository) GetConfiguration(ctx context.Context, name, version string) (model.Configuration, error) {
	key := m.makeConfigKey(name, version)
	config, exists := m.configs[key]
	if !exists {
		return model.Configuration{}, errors.New("configuration not found")
	}
	return config, nil
}

func (m *MockRepository) UpdateConfiguration(ctx context.Context, config model.Configuration) error {
	key := m.makeConfigKey(config.Name, config.Version)
	if _, exists := m.configs[key]; !exists {
		return errors.New("configuration not found")
	}
	m.configs[key] = config
	return nil
}

func (m *MockRepository) DeleteConfiguration(ctx context.Context, name, version string) error {
	key := m.makeConfigKey(name, version)
	if _, exists := m.configs[key]; !exists {
		return errors.New("configuration not found")
	}
	delete(m.configs, key)
	return nil
}

func (m *MockRepository) AddConfigurationGroup(ctx context.Context, group model.ConfigurationGroup) error {
	key := m.makeGroupKey(group.Name, group.Version)
	if _, exists := m.groups[key]; exists {
		return errors.New("configuration group already exists")
	}
	m.groups[key] = group
	return nil
}

func (m *MockRepository) GetConfigurationGroup(ctx context.Context, name, version string) (model.ConfigurationGroup, error) {
	key := m.makeGroupKey(name, version)
	group, exists := m.groups[key]
	if !exists {
		return model.ConfigurationGroup{}, errors.New("configuration group not found")
	}
	return group, nil
}

func (m *MockRepository) UpdateConfigurationGroup(ctx context.Context, group model.ConfigurationGroup) error {
	key := m.makeGroupKey(group.Name, group.Version)
	if _, exists := m.groups[key]; !exists {
		return errors.New("configuration group not found")
	}
	m.groups[key] = group
	return nil
}

func (m *MockRepository) DeleteConfigurationGroup(ctx context.Context, name, version string) error {
	key := m.makeGroupKey(name, version)
	if _, exists := m.groups[key]; !exists {
		return errors.New("configuration group not found")
	}
	delete(m.groups, key)
	return nil
}

// Tests
func TestConfigurationService_AddConfiguration(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewConfigurationService(mockRepo)
	ctx := context.Background()

	config := model.Configuration{
		ID:      uuid.New(),
		Name:    "test-service",
		Version: "v1.0.0",
		Params:  []model.Parameter{{Key: "port", Value: "8080"}},
	}

	// Test successful add
	err := service.AddConfiguration(ctx, config, "test-key-1")
	if err != nil {
		t.Fatalf("AddConfiguration failed: %v", err)
	}

	// Verify idempotency key was saved
	found, err := service.CheckIdempotencyKey(ctx, "test-key-1")
	if err != nil {
		t.Fatalf("CheckIdempotencyKey failed: %v", err)
	}
	if !found {
		t.Error("Idempotency key should have been saved")
	}

	// Test duplicate add
	err = service.AddConfiguration(ctx, config, "test-key-2")
	if err == nil {
		t.Error("Expected error for duplicate configuration")
	}
}

func TestConfigurationService_GetConfiguration(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewConfigurationService(mockRepo)
	ctx := context.Background()

	// First add a configuration
	config := model.Configuration{
		ID:      uuid.New(),
		Name:    "get-test",
		Version: "v1.0.0",
		Params:  []model.Parameter{{Key: "test", Value: "value"}},
	}

	err := service.AddConfiguration(ctx, config, "get-test-key")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Test successful get
	retrieved, err := service.GetConfiguration(ctx, "get-test", "v1.0.0")
	if err != nil {
		t.Fatalf("GetConfiguration failed: %v", err)
	}

	if retrieved.Name != "get-test" {
		t.Errorf("Expected name 'get-test', got '%s'", retrieved.Name)
	}

	// Test get non-existent
	_, err = service.GetConfiguration(ctx, "non-existent", "v1.0.0")
	if err == nil {
		t.Error("Expected error for non-existent configuration")
	}
}

func TestConfigurationService_UpdateConfiguration(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewConfigurationService(mockRepo)
	ctx := context.Background()

	// First add a configuration
	config := model.Configuration{
		ID:      uuid.New(),
		Name:    "update-test",
		Version: "v1.0.0",
		Params:  []model.Parameter{{Key: "old", Value: "value"}},
	}

	err := service.AddConfiguration(ctx, config, "update-test-key-1")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Test successful update
	updatedConfig := config
	updatedConfig.Params = []model.Parameter{{Key: "new", Value: "updated-value"}}

	err = service.UpdateConfiguration(ctx, updatedConfig, "update-test-key-2")
	if err != nil {
		t.Fatalf("UpdateConfiguration failed: %v", err)
	}

	// Verify idempotency key was saved
	found, err := service.CheckIdempotencyKey(ctx, "update-test-key-2")
	if err != nil {
		t.Fatalf("CheckIdempotencyKey failed: %v", err)
	}
	if !found {
		t.Error("Idempotency key should have been saved for update")
	}

	// Verify update
	retrieved, err := service.GetConfiguration(ctx, "update-test", "v1.0.0")
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}

	if len(retrieved.Params) != 1 || retrieved.Params[0].Key != "new" {
		t.Error("Configuration was not properly updated")
	}
}

func TestConfigurationService_Idempotency(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewConfigurationService(mockRepo)
	ctx := context.Background()

	// Test checking non-existent key
	found, err := service.CheckIdempotencyKey(ctx, "non-existent-key")
	if err != nil {
		t.Fatalf("CheckIdempotencyKey failed: %v", err)
	}
	if found {
		t.Error("Non-existent key should not be found")
	}

	// Test saving empty key (should not panic)
	service.SaveIdempotencyKey(ctx, "")
}

func TestConfigurationService_ConfigurationGroup(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewConfigurationService(mockRepo)
	ctx := context.Background()

	config := model.Configuration{
		ID:      uuid.New(),
		Name:    "group-config",
		Version: "v1.0.0",
		Params:  []model.Parameter{},
	}

	group := model.ConfigurationGroup{
		ID:             uuid.New(),
		Name:           "test-group",
		Version:        "v1.0.0",
		Configurations: []model.Configuration{config},
	}

	// Test add group
	err := service.AddConfigurationGroup(ctx, group, "group-key-1")
	if err != nil {
		t.Fatalf("AddConfigurationGroup failed: %v", err)
	}

	// Test get group
	retrieved, err := service.GetConfigurationGroup(ctx, "test-group", "v1.0.0")
	if err != nil {
		t.Fatalf("GetConfigurationGroup failed: %v", err)
	}

	if retrieved.Name != "test-group" {
		t.Errorf("Expected group name 'test-group', got '%s'", retrieved.Name)
	}
	if len(retrieved.Configurations) != 1 {
		t.Errorf("Expected 1 configuration in group, got %d", len(retrieved.Configurations))
	}
}
