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

	originalConfig, exists := m.configs[key]
	if !exists {
		return errors.New("configuration not found")
	}

	if config.ID == uuid.Nil {
		config.ID = originalConfig.ID
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
	originalGroup, exists := m.groups[key]

	if !exists {
		return errors.New("configuration group not found")
	}

	if group.ID == uuid.Nil {
		group.ID = originalGroup.ID
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

	err := service.AddConfiguration(ctx, config, "test-key-1")
	if err != nil {
		t.Fatalf("AddConfiguration failed: %v", err)
	}

	found, err := service.CheckIdempotencyKey(ctx, "test-key-1")
	if err != nil {
		t.Fatalf("CheckIdempotencyKey failed: %v", err)
	}
	if !found {
		t.Error("Idempotency key should have been saved")
	}

	err = service.AddConfiguration(ctx, config, "test-key-2")
	if err == nil {
		t.Error("Expected error for duplicate configuration")
	}
}

func TestConfigurationService_GetConfiguration(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewConfigurationService(mockRepo)
	ctx := context.Background()

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

	retrieved, err := service.GetConfiguration(ctx, "get-test", "v1.0.0")
	if err != nil {
		t.Fatalf("GetConfiguration failed: %v", err)
	}

	if retrieved.Name != "get-test" {
		t.Errorf("Expected name 'get-test', got '%s'", retrieved.Name)
	}

	_, err = service.GetConfiguration(ctx, "non-existent", "v1.0.0")
	if err == nil {
		t.Error("Expected error for non-existent configuration")
	}
}

func TestConfigurationService_UpdateConfiguration(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewConfigurationService(mockRepo)
	ctx := context.Background()

	originalID := uuid.New()
	config := model.Configuration{
		ID:      originalID,
		Name:    "update-test",
		Version: "v1.0.0",
		Params:  []model.Parameter{{Key: "old", Value: "value"}},
	}

	err := service.AddConfiguration(ctx, config, "update-test-key-1")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	updatedConfigInput := config
	updatedConfigInput.ID = uuid.Nil
	updatedConfigInput.Params = []model.Parameter{{Key: "new", Value: "updated-value"}}

	retConfig, err := service.UpdateConfiguration(ctx, updatedConfigInput, "update-test-key-2")
	if err != nil {
		t.Fatalf("UpdateConfiguration failed: %v", err)
	}

	if retConfig.ID != originalID {
		t.Errorf("Expected ID %s, but got ID %s in the returned configuration", originalID, retConfig.ID)
	}
	if len(retConfig.Params) != 1 || retConfig.Params[0].Key != "new" {
		t.Error("Returned configuration was not properly updated")
	}

	found, err := service.CheckIdempotencyKey(ctx, "update-test-key-2")
	if err != nil {
		t.Fatalf("CheckIdempotencyKey failed: %v", err)
	}
	if !found {
		t.Error("Idempotency key should have been saved for update")
	}

	nonExistentConfig := model.Configuration{Name: "non-exist", Version: "v1"}
	_, err = service.UpdateConfiguration(ctx, nonExistentConfig, "update-test-key-3")
	if err == nil {
		t.Error("Expected error for updating non-existent configuration")
	}
}

func TestConfigurationService_UpdateConfigurationGroup(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewConfigurationService(mockRepo)
	ctx := context.Background()

	originalID := uuid.New()
	originalConfig := model.Configuration{Params: []model.Parameter{{Key: "old", Value: "old-value"}}}
	group := model.ConfigurationGroup{
		ID:             originalID,
		Name:           "update-group-test",
		Version:        "v1.0.0",
		Configurations: []model.Configuration{originalConfig},
	}

	err := service.AddConfigurationGroup(ctx, group, "update-group-key-1")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	updatedConfig := model.Configuration{Params: []model.Parameter{{Key: "new", Value: "updated-value"}}}
	updatedGroupInput := group
	updatedGroupInput.ID = uuid.Nil
	updatedGroupInput.Configurations = []model.Configuration{updatedConfig}

	retGroup, err := service.UpdateConfigurationGroup(ctx, updatedGroupInput, "update-group-key-2")
	if err != nil {
		t.Fatalf("UpdateConfigurationGroup failed: %v", err)
	}

	if retGroup.ID != originalID {
		t.Errorf("Expected ID %s, but got ID %s in the returned group", originalID, retGroup.ID)
	}
	if len(retGroup.Configurations) != 1 || len(retGroup.Configurations[0].Params) != 1 || retGroup.Configurations[0].Params[0].Key != "new" {
		t.Errorf("Returned configuration group was not properly updated. Expected Key: 'new', got: %v", retGroup.Configurations[0].Params)
	}

	found, err := service.CheckIdempotencyKey(ctx, "update-group-key-2")
	if err != nil {
		t.Fatalf("CheckIdempotencyKey failed: %v", err)
	}
	if !found {
		t.Error("Idempotency key should have been saved for group update")
	}

	nonExistentGroup := model.ConfigurationGroup{Name: "non-exist-group", Version: "v1"}
	_, err = service.UpdateConfigurationGroup(ctx, nonExistentGroup, "update-group-key-3")
	if err == nil {
		t.Error("Expected error for updating non-existent group")
	}
}

func TestConfigurationService_Idempotency(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewConfigurationService(mockRepo)
	ctx := context.Background()

	found, err := service.CheckIdempotencyKey(ctx, "non-existent-key")
	if err != nil {
		t.Fatalf("CheckIdempotencyKey failed: %v", err)
	}
	if found {
		t.Error("Non-existent key should not be found")
	}

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
