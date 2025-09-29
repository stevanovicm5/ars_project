package services_test

import (
	"alati_projekat/model"
	"alati_projekat/services"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// --- MOCK REPOSITORY IMPLEMENTACIJA ---

type MockRepository struct {
	AddConfigCallCount    int
	SaveKeyCallCount      int
	GetConfigCallCount    int
	UpdateConfigCallCount int
	DeleteConfigCallCount int
	AddGroupCallCount     int

	MockAddConfiguration    func(config model.Configuration) error
	MockGetConfiguration    func(name, version string) (model.Configuration, error)
	MockUpdateConfiguration func(config model.Configuration) error
	MockDeleteConfiguration func(name, version string) error

	MockAddConfigurationGroup    func(group model.ConfigurationGroup) error
	MockGetConfigurationGroup    func(name, version string) (model.ConfigurationGroup, error)
	MockUpdateConfigurationGroup func(group model.ConfigurationGroup) error
	MockDeleteConfigurationGroup func(name, version string) error

	MockCheckKey func(key string) (bool, error)
	MockSaveKey  func(key string) error
}

// ------------------------------------------------
//    IMPLEMENTACIJA MOCK METODA
// ------------------------------------------------

func (m *MockRepository) AddConfiguration(config model.Configuration) error {
	m.AddConfigCallCount++
	if m.MockAddConfiguration != nil {
		return m.MockAddConfiguration(config)
	}
	return nil
}

func (m *MockRepository) GetConfiguration(name, version string) (model.Configuration, error) {
	m.GetConfigCallCount++
	if m.MockGetConfiguration != nil {
		return m.MockGetConfiguration(name, version)
	}
	return model.Configuration{}, errors.New("not implemented in mock")
}

func (m *MockRepository) UpdateConfiguration(config model.Configuration) error {
	m.UpdateConfigCallCount++
	if m.MockUpdateConfiguration != nil {
		return m.MockUpdateConfiguration(config)
	}
	return nil
}

func (m *MockRepository) DeleteConfiguration(name, version string) error {
	m.DeleteConfigCallCount++
	if m.MockDeleteConfiguration != nil {
		return m.MockDeleteConfiguration(name, version)
	}
	return nil
}

// --- CONFIGURATION GROUPS ---

func (m *MockRepository) AddConfigurationGroup(group model.ConfigurationGroup) error {
	m.AddGroupCallCount++
	if m.MockAddConfigurationGroup != nil {
		return m.MockAddConfigurationGroup(group)
	}
	return nil
}

func (m *MockRepository) GetConfigurationGroup(name, version string) (model.ConfigurationGroup, error) {
	if m.MockGetConfigurationGroup != nil {
		return m.MockGetConfigurationGroup(name, version)
	}
	return model.ConfigurationGroup{}, errors.New("not implemented in mock")
}

func (m *MockRepository) UpdateConfigurationGroup(group model.ConfigurationGroup) error {
	if m.MockUpdateConfigurationGroup != nil {
		return m.MockUpdateConfigurationGroup(group)
	}
	return nil
}

func (m *MockRepository) DeleteConfigurationGroup(name, version string) error {
	if m.MockDeleteConfigurationGroup != nil {
		return m.MockDeleteConfigurationGroup(name, version)
	}
	return nil
}

// --- IDEMPOTENCY ---

func (m *MockRepository) CheckIdempotencyKey(key string) (bool, error) {
	if m.MockCheckKey != nil {
		return m.MockCheckKey(key)
	}
	return false, nil
}

func (m *MockRepository) SaveIdempotencyKey(key string) error {
	m.SaveKeyCallCount++
	if m.MockSaveKey != nil {
		return m.MockSaveKey(key)
	}
	return nil
}

// ------------------------------------------------
//     TESTOVI SERVISNOG SLOJA
// ------------------------------------------------

// TEST CASE 1: Uspešno dodavanje konfiguracije
func TestAddConfiguration_Success(t *testing.T) {
	mockRepo := &MockRepository{
		MockAddConfiguration: func(config model.Configuration) error { return nil },
		MockSaveKey:          func(key string) error { return nil },
	}

	service := services.NewConfigurationService(mockRepo)

	testConfig := model.Configuration{ID: uuid.New(), Name: "Test", Version: "v1", Params: nil}
	key := "test-key-123"

	err := service.AddConfiguration(testConfig, key)

	if err != nil {
		t.Fatalf("Očekivana nil greška, dobijena: %v", err)
	}
	if mockRepo.AddConfigCallCount != 1 {
		t.Errorf("Očekivan 1 poziv AddConfiguration, dobijeno %d", mockRepo.AddConfigCallCount)
	}
	if mockRepo.SaveKeyCallCount != 1 {
		t.Errorf("Očekivan 1 poziv SaveIdempotencyKey, dobijeno %d", mockRepo.SaveKeyCallCount)
	}
}

// TEST CASE 2: Dodavanje ne uspe zbog greške Repozitorijuma (Rollback/Skip SaveKey)
func TestAddConfiguration_RepoError(t *testing.T) {
	expectedError := errors.New("Repo failed: conflict")

	mockRepo := &MockRepository{
		MockAddConfiguration: func(config model.Configuration) error { return expectedError },
		MockSaveKey:          func(key string) error { return nil },
	}

	service := services.NewConfigurationService(mockRepo)

	testConfig := model.Configuration{ID: uuid.New(), Name: "Test", Version: "v1", Params: nil}
	key := "test-key-456"

	err := service.AddConfiguration(testConfig, key)

	if err != expectedError {
		t.Errorf("Očekivana greška: %v, dobijena: %v", expectedError, err)
	}
	// Provera da li je SaveKey pozvan (NE SME BITI)
	if mockRepo.SaveKeyCallCount != 0 {
		t.Errorf("SaveIdempotencyKey NE SME biti pozvan ako AddConfiguration ne uspe, dobijeno %d poziva", mockRepo.SaveKeyCallCount)
	}
}

// TEST CASE 3: Provera da li Update poziva SaveKey
func TestUpdateConfiguration_Success(t *testing.T) {
	mockRepo := &MockRepository{
		MockUpdateConfiguration: func(config model.Configuration) error { return nil },
		MockSaveKey:             func(key string) error { return nil },
	}

	service := services.NewConfigurationService(mockRepo)

	testConfig := model.Configuration{ID: uuid.New(), Name: "Test", Version: "v1", Params: nil}
	key := "update-key-789"

	err := service.UpdateConfiguration(testConfig, key)

	if err != nil {
		t.Fatalf("Očekivana nil greška, dobijena: %v", err)
	}
	if mockRepo.UpdateConfigCallCount != 1 {
		t.Errorf("Očekivan 1 poziv UpdateConfiguration, dobijeno %d", mockRepo.UpdateConfigCallCount)
	}
	if mockRepo.SaveKeyCallCount != 1 {
		t.Errorf("Očekivan 1 poziv SaveIdempotencyKey, dobijeno %d", mockRepo.SaveKeyCallCount)
	}
}

// TEST CASE 4: Provera Idempotentnosti u servisu
func TestCheckIdempotencyKey_Processed(t *testing.T) {
	mockRepo := &MockRepository{
		MockCheckKey: func(key string) (bool, error) { return true, nil },
	}

	service := services.NewConfigurationService(mockRepo)
	key := "processed-key"

	isProcessed, err := service.CheckIdempotencyKey(key)

	if err != nil {
		t.Fatalf("Očekivana nil greška, dobijena: %v", err)
	}
	if !isProcessed {
		t.Errorf("Očekivano da je ključ obrađen (true), dobijeno: false")
	}
}
