package services

import (
	"alati_projekat/labels"
	"alati_projekat/model"
	"alati_projekat/repository"
	"context"
	"errors"
	"log"
	"strings"
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

func (s *ConfigurationService) CheckIdempotencyKey(ctx context.Context, key string) (bool, error) {
	return s.Repo.CheckIdempotencyKey(ctx, key)
}

func (s *ConfigurationService) SaveIdempotencyKey(ctx context.Context, key string) {
	if key == "" {
		return
	}
	if err := s.Repo.SaveIdempotencyKey(ctx, key); err != nil {
		log.Printf("IDEMPOTENCY WARNING: Failed to save key %s: %v", key, err)
	}
}

// --- CONFIGURATION CRUD LOGIC  ---

func (s *ConfigurationService) AddConfiguration(ctx context.Context, config model.Configuration, idempotencyKey string) error {
	if _, err := s.Repo.GetConfiguration(ctx, config.Name, config.Version); err == nil {
		return errors.New("configuration already exists (Conflict)")
	} else if !strings.Contains(err.Error(), "not found") {
		return err
	}

	if err := s.Repo.AddConfiguration(ctx, config); err != nil {
		return err
	}
	s.SaveIdempotencyKey(ctx, idempotencyKey)
	return nil
}

func (s *ConfigurationService) GetConfiguration(ctx context.Context, name string, version string) (model.Configuration, error) {
	return s.Repo.GetConfiguration(ctx, name, version)
}

func (s *ConfigurationService) UpdateConfiguration(ctx context.Context, config model.Configuration, idempotencyKey string) (model.Configuration, error) {
	existingConfig, err := s.Repo.GetConfiguration(ctx, config.Name, config.Version)
	if err != nil {
		return model.Configuration{}, err // Vraća "not found" ili drugu grešku
	}

	config.ID = existingConfig.ID

	if err := s.Repo.UpdateConfiguration(ctx, config); err != nil {
		return model.Configuration{}, err
	}
	s.SaveIdempotencyKey(ctx, idempotencyKey)
	return config, err
}

func (s *ConfigurationService) DeleteConfiguration(ctx context.Context, name string, version string) error {
	return s.Repo.DeleteConfiguration(ctx, name, version)
}

// --- CONFIGURATION GROUP CRUD LOGIC

func (s *ConfigurationService) AddConfigurationGroup(ctx context.Context, group model.ConfigurationGroup, idempotencyKey string) error {
	if _, err := s.Repo.GetConfigurationGroup(ctx, group.Name, group.Version); err == nil {
		return errors.New("configuration group already exists (Conflict)")
	} else if !strings.Contains(err.Error(), "not found") {
		return err
	}

	if err := s.Repo.AddConfigurationGroup(ctx, group); err != nil {
		return err
	}
	s.SaveIdempotencyKey(ctx, idempotencyKey)
	return nil
}

func (s *ConfigurationService) GetConfigurationGroup(ctx context.Context, name string, version string) (model.ConfigurationGroup, error) {
	return s.Repo.GetConfigurationGroup(ctx, name, version)
}

func (s *ConfigurationService) UpdateConfigurationGroup(ctx context.Context, group model.ConfigurationGroup, idempotencyKey string) (model.ConfigurationGroup, error) {
	existingGroup, err := s.Repo.GetConfigurationGroup(ctx, group.Name, group.Version)
	if err != nil {
		return model.ConfigurationGroup{}, err // Vraća "not found" ili drugu grešku
	}

	group.ID = existingGroup.ID

	if err := s.Repo.UpdateConfigurationGroup(ctx, group); err != nil {
		return model.ConfigurationGroup{}, err
	}
	s.SaveIdempotencyKey(ctx, idempotencyKey)
	return group, err
}

func (s *ConfigurationService) DeleteConfigurationGroup(ctx context.Context, name string, version string) error {
	return s.Repo.DeleteConfigurationGroup(ctx, name, version)
}

func (s *ConfigurationService) FilterConfigsByLabels(ctx context.Context, name, version string, want map[string]string) ([]model.Configuration, error) {
	g, err := s.Repo.GetConfigurationGroup(ctx, name, version)
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

func (s *ConfigurationService) DeleteConfigsByLabels(ctx context.Context, name, version string, want map[string]string) (int, error) {
	g, err := s.Repo.GetConfigurationGroup(ctx, name, version)
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
		return 0, nil
	}
	g.Configurations = filtered
	if err := s.Repo.AddConfigurationGroup(ctx, g); err != nil {
		return 0, err
	}
	return deleted, nil
}
