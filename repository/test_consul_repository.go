package repository

import (
	"alati_projekat/model"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

var consulRepo *ConsulRepository

func TestMain(m *testing.M) {
	consulAddr := "127.0.0.1:8500"

	repo, err := NewConsulRepository(consulAddr)
	if err != nil {
		log.Printf("SKIP INTEGRATION TESTS: Failed to connect to Consul at %s. Please run 'docker run -d --name=dev-consul -p 8500:8500 consul' before testing. Error: %v", consulAddr, err)
		os.Exit(0)
	}

	consulRepo = repo

	code := m.Run()
	os.Exit(code)
}

func cleanup(key string) {
	if consulRepo == nil || consulRepo.Client == nil {
		return
	}
	_, err := consulRepo.Client.KV().Delete(key, nil)
	if err != nil {
		log.Printf("Cleanup failed for key %s: %v", key, err)
	}
}

// ----------------------------------------------------
// TESTOVI ZA KONFIGURACIJE
// ----------------------------------------------------

func TestConsulConfigCRUD(t *testing.T) {
	name := "TestService"
	version := "v1.0"
	key := ConfigsPrefix + makeKey(name, version)

	defer cleanup(key)

	testConfig := model.Configuration{
		ID:      uuid.New(),
		Name:    name,
		Version: version,
		Params:  []model.Parameter{{Key: "db_url", Value: "localhost:5432"}},
	}

	// 1. CREATE (Add)
	t.Run("AddConfiguration", func(t *testing.T) {
		if err := consulRepo.AddConfiguration(testConfig); err != nil {
			t.Fatalf("FAIL: AddConfiguration failed: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	})

	// 2. READ (Get - uspešno)
	t.Run("GetConfiguration_Success", func(t *testing.T) {
		fetchedConfig, err := consulRepo.GetConfiguration(name, version)
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
		updatedConfig.Params = []model.Parameter{{Key: "db_url", Value: "remote:5432"}}

		if err := consulRepo.UpdateConfiguration(updatedConfig); err != nil {
			t.Fatalf("FAIL: UpdateConfiguration failed: %v", err)
		}

		checkConfig, _ := consulRepo.GetConfiguration(name, version)
		if checkConfig.Params[0].Value != "remote:5432" {
			t.Errorf("FAIL: Update failed, expected value remote:5432, got %s", checkConfig.Params[0].Value)
		}
	})

	// 4. DELETE (uspešno)
	t.Run("DeleteConfiguration_Success", func(t *testing.T) {
		if err := consulRepo.DeleteConfiguration(name, version); err != nil {
			t.Fatalf("FAIL: DeleteConfiguration failed: %v", err)
		}

		_, err := consulRepo.GetConfiguration(name, version)
		if err == nil {
			t.Fatalf("FAIL: Expected error after deletion, got nil")
		}
	})
}

// ----------------------------------------------------
// TESTOVI ZA KONFIGURACIONE GRUPE
// ----------------------------------------------------

func TestConsulGroupCRUD(t *testing.T) {
	name := "TestGroup"
	version := "v1.0"
	key := GroupsPrefix + makeKey(name, version)

	defer cleanup(key)

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
		if err := consulRepo.AddConfigurationGroup(testGroup); err != nil {
			t.Fatalf("FAIL: AddConfigurationGroup failed: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	})

	// 2. READ (Get Group - uspešno)
	t.Run("GetConfigurationGroup_Success", func(t *testing.T) {
		fetchedGroup, err := consulRepo.GetConfigurationGroup(name, version)
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

		if err := consulRepo.UpdateConfigurationGroup(updatedGroup); err != nil {
			t.Fatalf("FAIL: UpdateConfigurationGroup failed: %v", err)
		}

		checkGroup, _ := consulRepo.GetConfigurationGroup(name, version)
		if len(checkGroup.Configurations) != 2 {
			t.Errorf("FAIL: Update failed, expected 2 configurations, got %d", len(checkGroup.Configurations))
		}
	})

	// 4. DELETE (uspešno)
	t.Run("DeleteConfigurationGroup_Success", func(t *testing.T) {
		if err := consulRepo.DeleteConfigurationGroup(name, version); err != nil {
			t.Fatalf("FAIL: DeleteConfigurationGroup failed: %v", err)
		}

		_, err := consulRepo.GetConfigurationGroup(name, version)
		if err == nil {
			t.Fatalf("FAIL: Expected error after deletion, got nil")
		}
	})
}
