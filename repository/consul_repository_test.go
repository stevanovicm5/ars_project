package repository

import (
	"alati_projekat/model"
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestConsulRepository_ConfigurationCRUD(t *testing.T) {
	// Skip if Consul is not running (like in CI)
	repo, err := NewConsulRepository("http://localhost:8500")
	if err != nil {
		t.Skipf("Skipping test: Consul not available: %v", err)
	}

	ctx := context.Background()

	// Create unique test data to avoid conflicts
	testName := "test-config-" + uuid.New().String()[:8]
	testVersion := "v1.0.0"

	// Test data
	config := model.Configuration{
		ID:      uuid.New(),
		Name:    testName,
		Version: testVersion,
		Params: []model.Parameter{
			{Key: "database_url", Value: "postgres://localhost:5432/test"},
			{Key: "timeout", Value: "30s"},
		},
	}

	// Test 1: Add Configuration
	t.Run("AddConfiguration", func(t *testing.T) {
		err := repo.AddConfiguration(ctx, config)
		if err != nil {
			t.Fatalf("AddConfiguration failed: %v", err)
		}
	})

	// Test 2: Get Configuration
	t.Run("GetConfiguration", func(t *testing.T) {
		retrieved, err := repo.GetConfiguration(ctx, testName, testVersion)
		if err != nil {
			t.Fatalf("GetConfiguration failed: %v", err)
		}

		if retrieved.Name != testName {
			t.Errorf("Expected name %s, got %s", testName, retrieved.Name)
		}
		if retrieved.Version != testVersion {
			t.Errorf("Expected version %s, got %s", testVersion, retrieved.Version)
		}
		if len(retrieved.Params) != 2 {
			t.Errorf("Expected 2 params, got %d", len(retrieved.Params))
		}
	})

	// Test 3: Update Configuration
	t.Run("UpdateConfiguration", func(t *testing.T) {
		updatedConfig := config
		updatedConfig.Params = append(updatedConfig.Params, model.Parameter{
			Key: "new_setting", Value: "new_value",
		})

		err := repo.UpdateConfiguration(ctx, updatedConfig)
		if err != nil {
			t.Fatalf("UpdateConfiguration failed: %v", err)
		}

		// Verify update
		retrieved, err := repo.GetConfiguration(ctx, testName, testVersion)
		if err != nil {
			t.Fatalf("GetConfiguration after update failed: %v", err)
		}

		if len(retrieved.Params) != 3 {
			t.Errorf("Expected 3 params after update, got %d", len(retrieved.Params))
		}
	})

	// Test 4: Delete Configuration
	t.Run("DeleteConfiguration", func(t *testing.T) {
		err := repo.DeleteConfiguration(ctx, testName, testVersion)
		if err != nil {
			t.Fatalf("DeleteConfiguration failed: %v", err)
		}

		// Verify deletion
		_, err = repo.GetConfiguration(ctx, testName, testVersion)
		if err == nil {
			t.Error("Expected error after deletion, but got none")
		}
		if !contains(err.Error(), "not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})
}

func TestConsulRepository_ConfigurationGroupCRUD(t *testing.T) {
	repo, err := NewConsulRepository("http://localhost:8500")
	if err != nil {
		t.Skipf("Skipping test: Consul not available: %v", err)
	}

	ctx := context.Background()

	// Create unique test data to avoid conflicts
	testName := "test-group-" + uuid.New().String()[:8]
	testVersion := "v1.0.0"

	// Test data for configuration group
	group := model.ConfigurationGroup{
		ID:      uuid.New(),
		Name:    testName,
		Version: testVersion,
		Configurations: []model.Configuration{
			{
				ID:      uuid.New(),
				Name:    "config1",
				Version: "v1.0.0",
				Params: []model.Parameter{
					{Key: "key1", Value: "value1"},
				},
			},
			{
				ID:      uuid.New(),
				Name:    "config2",
				Version: "v1.0.0",
				Params: []model.Parameter{
					{Key: "key2", Value: "value2"},
				},
			},
		},
	}

	// Test 1: Add Configuration Group
	t.Run("AddConfigurationGroup", func(t *testing.T) {
		err := repo.AddConfigurationGroup(ctx, group)
		if err != nil {
			t.Fatalf("AddConfigurationGroup failed: %v", err)
		}
	})

	// Test 2: Get Configuration Group
	t.Run("GetConfigurationGroup", func(t *testing.T) {
		retrieved, err := repo.GetConfigurationGroup(ctx, testName, testVersion)
		if err != nil {
			t.Fatalf("GetConfigurationGroup failed: %v", err)
		}

		if retrieved.Name != testName {
			t.Errorf("Expected name %s, got %s", testName, retrieved.Name)
		}
		if retrieved.Version != testVersion {
			t.Errorf("Expected version %s, got %s", testVersion, retrieved.Version)
		}
		if len(retrieved.Configurations) != 2 {
			t.Errorf("Expected 2 configurations in group, got %d", len(retrieved.Configurations))
		}
	})

	// Test 3: Update Configuration Group
	t.Run("UpdateConfigurationGroup", func(t *testing.T) {
		updatedGroup := group
		updatedGroup.Configurations = append(updatedGroup.Configurations, model.Configuration{
			ID:      uuid.New(),
			Name:    "config3",
			Version: "v1.0.0",
			Params: []model.Parameter{
				{Key: "key3", Value: "value3"},
			},
		})

		err := repo.UpdateConfigurationGroup(ctx, updatedGroup)
		if err != nil {
			t.Fatalf("UpdateConfigurationGroup failed: %v", err)
		}

		// Verify update
		retrieved, err := repo.GetConfigurationGroup(ctx, testName, testVersion)
		if err != nil {
			t.Fatalf("GetConfigurationGroup after update failed: %v", err)
		}

		if len(retrieved.Configurations) != 3 {
			t.Errorf("Expected 3 configurations after update, got %d", len(retrieved.Configurations))
		}
	})

	// Test 4: Delete Configuration Group
	t.Run("DeleteConfigurationGroup", func(t *testing.T) {
		err := repo.DeleteConfigurationGroup(ctx, testName, testVersion)
		if err != nil {
			t.Fatalf("DeleteConfigurationGroup failed: %v", err)
		}

		// Verify deletion
		_, err = repo.GetConfigurationGroup(ctx, testName, testVersion)
		if err == nil {
			t.Error("Expected error after deletion, but got none")
		}
		if !contains(err.Error(), "not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})
}

func TestConsulRepository_Idempotency(t *testing.T) {
	repo, err := NewConsulRepository("http://localhost:8500")
	if err != nil {
		t.Skipf("Skipping test: Consul not available: %v", err)
	}

	ctx := context.Background()
	testKey := "test-idempotency-" + uuid.New().String()

	t.Run("CheckNonExistentKey", func(t *testing.T) {
		found, err := repo.CheckIdempotencyKey(ctx, testKey)
		if err != nil {
			t.Fatalf("CheckIdempotencyKey failed: %v", err)
		}
		if found {
			t.Error("Expected key to not be found")
		}
	})

	t.Run("SaveAndCheckKey", func(t *testing.T) {
		err := repo.SaveIdempotencyKey(ctx, testKey)
		if err != nil {
			t.Fatalf("SaveIdempotencyKey failed: %v", err)
		}

		found, err := repo.CheckIdempotencyKey(ctx, testKey)
		if err != nil {
			t.Fatalf("CheckIdempotencyKey after save failed: %v", err)
		}
		if !found {
			t.Error("Expected key to be found after save")
		}
	})

	// Cleanup
	_ = repo.SaveIdempotencyKey(ctx, testKey) // This will overwrite, but that's fine for test cleanup
}

func TestConsulRepository_GetNonExistentConfiguration(t *testing.T) {
	repo, err := NewConsulRepository("http://localhost:8500")
	if err != nil {
		t.Skipf("Skipping test: Consul not available: %v", err)
	}

	ctx := context.Background()

	_, err = repo.GetConfiguration(ctx, "non-existent-config", "v999.0.0")
	if err == nil {
		t.Error("Expected error for non-existent configuration")
	}
	if !contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

func TestConsulRepository_GetNonExistentConfigurationGroup(t *testing.T) {
	repo, err := NewConsulRepository("http://localhost:8500")
	if err != nil {
		t.Skipf("Skipping test: Consul not available: %v", err)
	}

	ctx := context.Background()

	_, err = repo.GetConfigurationGroup(ctx, "non-existent-group", "v999.0.0")
	if err == nil {
		t.Error("Expected error for non-existent configuration group")
	}
	if !contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}
