package services

import (
	"alati_projekat/labels"
	"alati_projekat/model"
	"alati_projekat/repository"
	"log"
)

type ConfigurationService struct {
	Repo repository.Repository
}

func NewConfigurationService(repo repository.Repository) *ConfigurationService {
	return &ConfigurationService{
		Repo: repo,
	}
}

var _ Service = (*ConfigurationService)(nil)

// --- IDEMPOTENCY LOGIC ---

func (s *ConfigurationService) CheckIdempotencyKey(key string) (bool, error) {
	return s.Repo.CheckIdempotencyKey(key)
}

func (s *ConfigurationService) SaveIdempotencyKey(key string) {
	if key == "" {
		return
	}
	if err := s.Repo.SaveIdempotencyKey(key); err != nil {
		log.Printf("IDEMPOTENCY WARNING: Failed to save key %s: %v", key, err)
	}
}

// --- CONFIGURATION CRUD LOGIC  ---

func (s *ConfigurationService) AddConfiguration(config model.Configuration, idempotencyKey string) error {
	if err := s.Repo.AddConfiguration(config); err != nil {
		return err
	}
	s.SaveIdempotencyKey(idempotencyKey)
	return nil
}

func (s *ConfigurationService) GetConfiguration(name string, version string) (model.Configuration, error) {
	return s.Repo.GetConfiguration(name, version)
}

func (s *ConfigurationService) UpdateConfiguration(config model.Configuration, idempotencyKey string) error {
	if err := s.Repo.UpdateConfiguration(config); err != nil {
		return err
	}
	s.SaveIdempotencyKey(idempotencyKey)
	return nil
}

func (s *ConfigurationService) DeleteConfiguration(name string, version string) error {
	return s.Repo.DeleteConfiguration(name, version)
}

// --- CONFIGURATION GROUP CRUD LOGIC

func (s *ConfigurationService) AddConfigurationGroup(group model.ConfigurationGroup, idempotencyKey string) error {
	if err := s.Repo.AddConfigurationGroup(group); err != nil {
		return err
	}
	s.SaveIdempotencyKey(idempotencyKey)
	return nil
}

func (s *ConfigurationService) GetConfigurationGroup(name string, version string) (model.ConfigurationGroup, error) {
	return s.Repo.GetConfigurationGroup(name, version)
}

func (s *ConfigurationService) UpdateConfigurationGroup(group model.ConfigurationGroup, idempotencyKey string) error {
	if err := s.Repo.UpdateConfigurationGroup(group); err != nil {
		return err
	}
	s.SaveIdempotencyKey(idempotencyKey)
	return nil
}

func (s *ConfigurationService) DeleteConfigurationGroup(name string, version string) error {
	return s.Repo.DeleteConfigurationGroup(name, version)
}

func (s *ConfigurationService) FilterConfigsByLabels(name, version string, want map[string]string) ([]model.Configuration, error) {
	g, err := s.Repo.GetConfigurationGroup(name, version)
	if err != nil {
		return nil, err
	}
	var out []model.Configuration
	for _, cfg := range g.Configurations {
		if labels.HasAll(cfg, want) {
			out = append(out, cfg)
		}
	}
	return out, nil
}

func (s *ConfigurationService) DeleteConfigsByLabels(name, version string, want map[string]string) (int, error) {
	g, err := s.Repo.GetConfigurationGroup(name, version)
	if err != nil {
		return 0, err
	}
	filtered := make([]model.Configuration, 0, len(g.Configurations))
	deleted := 0
	for _, cfg := range g.Configurations {
		if labels.HasAll(cfg, want) {
			deleted++
			continue
		}
		filtered = append(filtered, cfg)
	}
	if deleted == 0 {
		// nije greška – samo niko nije pogođen
		return 0, nil
	}
	g.Configurations = filtered
	if err := s.Repo.AddConfigurationGroup(g); err != nil {
		return 0, err
	}
	return deleted, nil
}
