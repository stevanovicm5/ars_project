package repository

import (
	"alati_projekat/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/consul/api"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const (
	ConfigsPrefix = "configurations/"
	GroupsPrefix  = "configgroups/"
)

var tracer = otel.Tracer("consul-repository-tracer")

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

func makeKey(name, version string) string {
	return fmt.Sprintf("%s/%s", name, version)
}

// ---------------------- CONFIGURATIONS ----------------------

func (r *ConsulRepository) AddConfiguration(ctx context.Context, config model.Configuration) (err error) {
	ctx, span := tracer.Start(ctx, "AddConfiguration")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()
	span.SetAttributes(attribute.String("config.name", config.Name), attribute.String("config.version", config.Version))

	key := ConfigsPrefix + makeKey(config.Name, config.Version)
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize configuration: %w", err)
	}

	p := &api.KVPair{Key: key, Value: data}

	writeOptions := (&api.WriteOptions{}).WithContext(ctx)

	_, err = r.Client.KV().Put(p, writeOptions)
	if err != nil {
		return fmt.Errorf("failed to put configuration into Consul: %w", err)
	}

	return nil
}

func (r *ConsulRepository) GetConfiguration(ctx context.Context, name, version string) (config model.Configuration, err error) {
	ctx, span := tracer.Start(ctx, "GetConfiguration")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()
	span.SetAttributes(attribute.String("config.name", name), attribute.String("config.version", version))

	key := ConfigsPrefix + makeKey(name, version)

	queryOptions := (&api.QueryOptions{}).WithContext(ctx)

	pair, _, err := r.Client.KV().Get(key, queryOptions)
	if err != nil {
		return model.Configuration{}, fmt.Errorf("failed to get configuration from Consul: %w", err)
	}

	if pair == nil {
		return model.Configuration{}, errors.New("configuration not found")
	}

	if err := json.Unmarshal(pair.Value, &config); err != nil {
		return model.Configuration{}, fmt.Errorf("failed to decode configuration JSON: %w", err)
	}

	return config, nil
}

func (r *ConsulRepository) UpdateConfiguration(ctx context.Context, config model.Configuration) (err error) {
	ctx, span := tracer.Start(ctx, "UpdateConfiguration")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()
	span.SetAttributes(attribute.String("config.name", config.Name), attribute.String("config.version", config.Version))

	if _, err := r.GetConfiguration(ctx, config.Name, config.Version); err != nil {
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

	writeOptions := (&api.WriteOptions{}).WithContext(ctx)

	_, err = r.Client.KV().Put(p, writeOptions)
	if err != nil {
		return fmt.Errorf("failed to update configuration in Consul: %w", err)
	}

	return nil
}

func (r *ConsulRepository) DeleteConfiguration(ctx context.Context, name, version string) (err error) {
	ctx, span := tracer.Start(ctx, "DeleteConfiguration")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()
	span.SetAttributes(attribute.String("config.name", name), attribute.String("config.version", version))

	key := ConfigsPrefix + makeKey(name, version)

	writeOptions := (&api.WriteOptions{}).WithContext(ctx)

	_, err = r.Client.KV().Delete(key, writeOptions)
	if err != nil {
		return fmt.Errorf("failed to delete configuration from Consul: %w", err)
	}

	return nil
}

// ---------------------- CONFIGURATION GROUPS ----------------------

func (r *ConsulRepository) AddConfigurationGroup(ctx context.Context, group model.ConfigurationGroup) (err error) {
	ctx, span := tracer.Start(ctx, "AddConfigurationGroup")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()
	span.SetAttributes(attribute.String("group.name", group.Name), attribute.String("group.version", group.Version))

	key := GroupsPrefix + makeKey(group.Name, group.Version)
	data, err := json.Marshal(group)
	if err != nil {
		return fmt.Errorf("failed to serialize configuration group: %w", err)
	}

	p := &api.KVPair{Key: key, Value: data}

	writeOptions := (&api.WriteOptions{}).WithContext(ctx)

	_, err = r.Client.KV().Put(p, writeOptions)
	if err != nil {
		return fmt.Errorf("failed to put configuration group into Consul: %w", err)
	}

	return nil
}

func (r *ConsulRepository) GetConfigurationGroup(ctx context.Context, name, version string) (group model.ConfigurationGroup, err error) {
	ctx, span := tracer.Start(ctx, "GetConfigurationGroup")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()
	span.SetAttributes(attribute.String("group.name", name), attribute.String("group.version", version))

	key := GroupsPrefix + makeKey(name, version)

	queryOptions := (&api.QueryOptions{}).WithContext(ctx)

	pair, _, err := r.Client.KV().Get(key, queryOptions)
	if err != nil {
		return model.ConfigurationGroup{}, fmt.Errorf("failed to get configuration group from Consul: %w", err)
	}

	if pair == nil {
		return model.ConfigurationGroup{}, errors.New("configuration group not found")
	}

	if err := json.Unmarshal(pair.Value, &group); err != nil {
		return model.ConfigurationGroup{}, fmt.Errorf("failed to decode configuration group JSON: %w", err)
	}

	return group, nil
}

func (r *ConsulRepository) UpdateConfigurationGroup(ctx context.Context, group model.ConfigurationGroup) (err error) {
	ctx, span := tracer.Start(ctx, "UpdateConfigurationGroup")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()
	span.SetAttributes(attribute.String("group.name", group.Name), attribute.String("group.version", group.Version))

	if _, err := r.GetConfigurationGroup(ctx, group.Name, group.Version); err != nil {
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

	writeOptions := (&api.WriteOptions{}).WithContext(ctx)

	_, err = r.Client.KV().Put(p, writeOptions)
	if err != nil {
		return fmt.Errorf("failed to update configuration group in Consul: %w", err)
	}

	return nil
}

func (r *ConsulRepository) DeleteConfigurationGroup(ctx context.Context, name, version string) (err error) {
	ctx, span := tracer.Start(ctx, "DeleteConfigurationGroup")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()
	span.SetAttributes(attribute.String("group.name", name), attribute.String("group.version", version))

	key := GroupsPrefix + makeKey(name, version)

	writeOptions := (&api.WriteOptions{}).WithContext(ctx)

	_, err = r.Client.KV().Delete(key, writeOptions)
	if err != nil {
		return fmt.Errorf("failed to delete configuration group from Consul: %w", err)
	}

	return nil
}

// ---------------------- IDEMPOTENCY ----------------------

const IdempotencyPrefix = "idempotency/"

func (r *ConsulRepository) CheckIdempotencyKey(ctx context.Context, key string) (found bool, err error) {
	ctx, span := tracer.Start(ctx, "CheckIdempotencyKey")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()
	span.SetAttributes(attribute.String("idempotency.key", key))

	fullKey := IdempotencyPrefix + key

	queryOptions := (&api.QueryOptions{}).WithContext(ctx)

	pair, _, err := r.Client.KV().Get(fullKey, queryOptions)
	if err != nil {
		return false, fmt.Errorf("consul check failed: %w", err)
	}
	return pair != nil, nil
}

func (r *ConsulRepository) SaveIdempotencyKey(ctx context.Context, key string) (err error) {
	ctx, span := tracer.Start(ctx, "SaveIdempotencyKey")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()
	span.SetAttributes(attribute.String("idempotency.key", key))

	fullKey := IdempotencyPrefix + key
	p := &api.KVPair{Key: fullKey, Value: []byte("processed")}

	writeOptions := (&api.WriteOptions{}).WithContext(ctx)

	_, err = r.Client.KV().Put(p, writeOptions)
	if err != nil {
		return fmt.Errorf("failed to save idempotency key to Consul: %w", err)
	}
	return nil
}
