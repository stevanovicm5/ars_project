package repository

import (
	"alati_projekat/model"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/consul/api"
)

const (
	ConfigsPrefix = "configurations/"
	GroupsPrefix  = "configgroups/"
)

type ConsulRepository struct {
	Client *api.Client
}

func NewConsulRepository(addr string) (*ConsulRepository, error) {

	config := api.DefaultConfig()
	config.Address = addr

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Consul client: %w", err)
	}

	return &ConsulRepository{
		Client: client,
	}, nil
}

// CONFIGURATIONS
// ADD
func (r *ConsulRepository) AddConfiguration(config model.Configuration) error {
	key := ConfigsPrefix + makeKey(config.Name, config.Version)

	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize configuration: %w", err)
	}

	p := &api.KVPair{Key: key, Value: data}

	_, err = r.Client.KV().Put(p, nil)
	if err != nil {
		return fmt.Errorf("failed to put configuration into Consul: %w", err)
	}

	return nil
}

// GET
func (r *ConsulRepository) GetConfiguration(name, version string) (model.Configuration, error) {
	// ISPRAVKA: Koristimo ConfigsPrefix
	key := ConfigsPrefix + makeKey(name, version)

	pair, _, err := r.Client.KV().Get(key, nil)
	if err != nil {
		return model.Configuration{}, fmt.Errorf("failed to get configuration from Consul: %w", err)
	}

	if pair == nil {
		return model.Configuration{}, errors.New("configuration not found")
	}

	var config model.Configuration
	if err := json.Unmarshal(pair.Value, &config); err != nil {
		return model.Configuration{}, fmt.Errorf("failed to decode configuration JSON: %w", err)
	}

	return config, nil
}

// UPDATE
func (r *ConsulRepository) UpdateConfiguration(config model.Configuration) error {
	_, err := r.GetConfiguration(config.Name, config.Version)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return errors.New("configuration not found for update")
		}
		return err
	}

	key := ConfigsPrefix + makeKey(config.Name, config.Version)

	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize configuration for update: %w", err)
	}

	p := &api.KVPair{Key: key, Value: data}
	_, err = r.Client.KV().Put(p, nil)
	if err != nil {
		return fmt.Errorf("failed to update configuration in Consul: %w", err)
	}

	return nil
}

// DELETE
func (r *ConsulRepository) DeleteConfiguration(name, version string) error {
	key := ConfigsPrefix + makeKey(name, version)

	_, err := r.Client.KV().Delete(key, nil)
	if err != nil {
		return fmt.Errorf("failed to delete configuration from Consul: %w", err)
	}

	return nil
}

// CONFIGURATION GROUPS
// ADD
func (r *ConsulRepository) AddConfigurationGroup(group model.ConfigurationGroup) error {
	key := GroupsPrefix + makeKey(group.Name, group.Version)

	data, err := json.Marshal(group)
	if err != nil {
		return fmt.Errorf("failed to serialize configuration group: %w", err)
	}

	p := &api.KVPair{Key: key, Value: data}

	_, err = r.Client.KV().Put(p, nil)
	if err != nil {
		return fmt.Errorf("failed to put configuration group into Consul: %w", err)
	}

	return nil
}

// GET
func (r *ConsulRepository) GetConfigurationGroup(name, version string) (model.ConfigurationGroup, error) {
	key := GroupsPrefix + makeKey(name, version)

	pair, _, err := r.Client.KV().Get(key, nil)
	if err != nil {
		return model.ConfigurationGroup{}, fmt.Errorf("failed to get configuration group from Consul: %w", err)
	}

	if pair == nil {
		return model.ConfigurationGroup{}, errors.New("configuration group not found")
	}

	var group model.ConfigurationGroup
	if err := json.Unmarshal(pair.Value, &group); err != nil {
		return model.ConfigurationGroup{}, fmt.Errorf("failed to decode configuration group JSON: %w", err)
	}

	return group, nil
}

// UPDATE
func (r *ConsulRepository) UpdateConfigurationGroup(group model.ConfigurationGroup) error {
	_, err := r.GetConfigurationGroup(group.Name, group.Version)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return errors.New("configuration group not found for update")
		}
		return err
	}

	key := GroupsPrefix + makeKey(group.Name, group.Version)

	data, err := json.Marshal(group)
	if err != nil {
		return fmt.Errorf("failed to serialize configuration group for update: %w", err)
	}

	p := &api.KVPair{Key: key, Value: data}
	_, err = r.Client.KV().Put(p, nil)
	if err != nil {
		return fmt.Errorf("failed to update configuration group in Consul: %w", err)
	}

	return nil
}

// DELETE
func (r *ConsulRepository) DeleteConfigurationGroup(name, version string) error {
	key := GroupsPrefix + makeKey(name, version)

	_, err := r.Client.KV().Delete(key, nil)
	if err != nil {
		return fmt.Errorf("failed to delete configuration group from Consul: %w", err)
	}

	return nil
}

const IdempotencyPrefix = "idempotency/"

func (r *ConsulRepository) CheckIdempotencyKey(key string) (bool, error) {
	fullKey := IdempotencyPrefix + key
	pair, _, err := r.Client.KV().Get(fullKey, nil)
	if err != nil {
		return false, fmt.Errorf("consul check failed: %w", err)
	}
	return pair != nil, nil
}

func (r *ConsulRepository) SaveIdempotencyKey(key string) error {
	fullKey := IdempotencyPrefix + key
	p := &api.KVPair{Key: fullKey, Value: []byte("processed")}

	_, err := r.Client.KV().Put(p, nil)
	if err != nil {
		return fmt.Errorf("failed to save idempotency key to Consul: %w", err)
	}
	return nil
}
