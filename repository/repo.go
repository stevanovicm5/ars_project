package repository

import (
	"alati_projekat/model"
	"errors"
	"fmt"
)

// Define interface for the repository to allow future switching
type ConfigRepository interface {
	AddConfiguration(config model.Configuration) error
	GetConfiguration(name, version string) (model.Configuration, error)
	DeleteConfiguration(name, version string) error

	AddConfigurationGroup(model.ConfigurationGroup) error
	GetConfigurationGroup(name, version string) (model.ConfigurationGroup, error)
	DeleteConfigurationGroup(name, version string) error
}

type InMemoryRepository struct {
	configs map[string]model.Configuration
	groups  map[string]model.ConfigurationGroup
}

func makeKey(name, version string) string {
	return fmt.Sprintf("%s#%s", name, version)
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		configs: make(map[string]model.Configuration),
		groups:  make(map[string]model.ConfigurationGroup),
	}
}

func (r *InMemoryRepository) AddConfiguration(config model.Configuration) error {
	key := makeKey(config.Name, config.Version)

	if _, exists := r.configs[key]; exists {
		return errors.New("configuration with this name and version already exists")
	}
	r.configs[key] = config
	return nil
}

func (r *InMemoryRepository) GetConfiguration(name, version string) (model.Configuration, error) {
	key := makeKey(name, version)
	config, exists := r.configs[key]
	if !exists {
		return model.Configuration{}, errors.New("configuration not found")
	}
	return config, nil
}

func (r *InMemoryRepository) DeleteConfiguration(name, version string) error {
	key := makeKey(name, version)
	if _, exists := r.configs[key]; !exists {
		return errors.New("configuration not found for deletion")
	}
	delete(r.configs, key)
	return nil
}

func (r *InMemoryRepository) AddConfigurationGroup(group model.ConfigurationGroup) error {
	key := makeKey(group.Name, group.Version)
	if _, exists := r.groups[key]; exists {
		return errors.New("config group with this name and version already exists")
	}
	r.groups[key] = group
	return nil
}

func (r *InMemoryRepository) GetConfigurationGroup(name, version string) (model.ConfigurationGroup, error) {
	key := makeKey(name, version)
	group, exists := r.groups[key]
	if !exists {
		return model.ConfigurationGroup{}, errors.New("config group not found")
	}
	return group, nil
}

func (r *InMemoryRepository) DeleteConfigurationGroup(name, version string) error {
	key := makeKey(name, version)
	if _, exists := r.groups[key]; !exists {
		return errors.New("config group not found for deletion")
	}
	delete(r.groups, key)
	return nil
}
