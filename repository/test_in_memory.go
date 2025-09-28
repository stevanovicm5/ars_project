package repository

import (
	"alati_projekat/model"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func setupRepo() *InMemoryRepository {
	return NewInMemoryRepository()
}

// ----------------------------------------------------
// TESTOVI ZA KONFIGURACIJE
// ----------------------------------------------------

func TestConfigCRUD(t *testing.T) {
	repo := setupRepo()

	name := "ServiceA"
	version := "v1.0.0"
	key := makeKey(name, version)

	testConfig := model.Configuration{
		ID:      uuid.New(),
		Name:    name,
		Version: version,
		Params:  []model.Parameter{{Key: "timeout", Value: "10s"}},
	}

	// 1. CREATE (Add)
	t.Run("AddConfiguration", func(t *testing.T) {
		if err := repo.AddConfiguration(testConfig); err != nil {
			t.Fatalf("FAIL: AddConfiguration should succeed, got error: %v", err)
		}
		if _, exists := repo.configs[key]; !exists {
			t.Fatal("FAIL: Configuration not found in map after Add")
		}
	})

	// 1. CREATE (Duplikat)
	t.Run("AddDuplicateConfiguration", func(t *testing.T) {
		err := repo.AddConfiguration(testConfig)
		if err == nil || !strings.Contains(err.Error(), "already exists") {
			t.Errorf("FAIL: Expected 'already exists' error, got: %v", err)
		}
	})

	// 2. READ (Get - uspešno)
	t.Run("GetConfiguration_Success", func(t *testing.T) {
		fetchedConfig, err := repo.GetConfiguration(name, version)
		if err != nil {
			t.Fatalf("FAIL: GetConfiguration failed: %v", err)
		}
		if fetchedConfig.Name != name {
			t.Errorf("FAIL: Expected name %s, got %s", name, fetchedConfig.Name)
		}
	})

	// 3. UPDATE (Put)
	t.Run("UpdateConfiguration_Success", func(t *testing.T) {
		updatedConfig := testConfig
		updatedConfig.Params = []model.Parameter{{Key: "timeout", Value: "20s"}}

		if err := repo.UpdateConfiguration(updatedConfig); err != nil {
			t.Fatalf("FAIL: UpdateConfiguration failed: %v", err)
		}

		checkConfig, _ := repo.GetConfiguration(name, version)
		if checkConfig.Params[0].Value != "20s" {
			t.Errorf("FAIL: Update failed, expected value 20s, got %s", checkConfig.Params[0].Value)
		}
	})

	// 4. DELETE (uspešno)
	t.Run("DeleteConfiguration_Success", func(t *testing.T) {
		if err := repo.DeleteConfiguration(name, version); err != nil {
			t.Fatalf("FAIL: DeleteConfiguration failed: %v", err)
		}

		_, err := repo.GetConfiguration(name, version)
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("FAIL: Expected 'not found' error after deletion, got: %v", err)
		}
	})

	// 4. DELETE (not found)
	t.Run("DeleteConfiguration_NotFound", func(t *testing.T) {
		err := repo.DeleteConfiguration(name, version)
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("FAIL: Expected 'not found' error on second deletion attempt, got: %v", err)
		}
	})
}

// ----------------------------------------------------
// TESTOVI ZA GRUPE
// ----------------------------------------------------

func TestGroupCRUD(t *testing.T) {
	repo := setupRepo()

	name := "WebServers"
	version := "v2.0.0"
	key := makeKey(name, version)

	testGroup := model.ConfigurationGroup{
		ID:      uuid.New(),
		Name:    name,
		Version: version,
		Configurations: []model.Configuration{
			{Name: "C1", Version: "v1"},
		},
	}

	// 1. CREATE (Add Group)
	t.Run("AddConfigurationGroup", func(t *testing.T) {
		if err := repo.AddConfigurationGroup(testGroup); err != nil {
			t.Fatalf("FAIL: AddConfigurationGroup should succeed, got error: %v", err)
		}
		if _, exists := repo.groups[key]; !exists {
			t.Fatal("FAIL: Group not found in map after Add")
		}
	})

	// 2. READ (Get Group - uspešno)
	t.Run("GetConfigurationGroup_Success", func(t *testing.T) {
		fetchedGroup, err := repo.GetConfigurationGroup(name, version)
		if err != nil {
			t.Fatalf("FAIL: GetConfigurationGroup failed: %v", err)
		}
		if fetchedGroup.Name != name {
			t.Errorf("FAIL: Expected name %s, got %s", name, fetchedGroup.Name)
		}
	})

	// 3. UPDATE (Put Group)
	t.Run("UpdateConfigurationGroup_Success", func(t *testing.T) {
		updatedGroup := testGroup
		updatedGroup.Configurations = append(updatedGroup.Configurations, model.Configuration{Name: "C2", Version: "v1"})

		if err := repo.UpdateConfigurationGroup(updatedGroup); err != nil {
			t.Fatalf("FAIL: UpdateConfigurationGroup failed: %v", err)
		}

		checkGroup, _ := repo.GetConfigurationGroup(name, version)
		if len(checkGroup.Configurations) != 2 {
			t.Errorf("FAIL: Update failed, expected 2 configurations, got %d", len(checkGroup.Configurations))
		}
	})

	// 4. DELETE (uspešno)
	t.Run("DeleteConfigurationGroup_Success", func(t *testing.T) {
		if err := repo.DeleteConfigurationGroup(name, version); err != nil {
			t.Fatalf("FAIL: DeleteConfigurationGroup failed: %v", err)
		}

		_, err := repo.GetConfigurationGroup(name, version)
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("FAIL: Expected 'not found' error after deletion, got: %v", err)
		}
	})
}
