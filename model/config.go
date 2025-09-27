package model

import "github.com/google/uuid"

// Parameter represents a key-value pair within a configuration or a label.
type Parameter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Configuration represents a single configuration item.
type Configuration struct {
	ID      uuid.UUID   `json:"id"`
	Name    string      `json:"name"`
	Version string      `json:"version"`
	Params  []Parameter `json:"params"`
}

// ConfigurationGroup represents a collection of configurations.
type ConfigurationGroup struct {
	ID             uuid.UUID       `json:"id"`
	Name           string          `json:"name"`
	Version        string          `json:"version"`
	Configurations []Configuration `json:"configurations"`
}

// Input structs: We often define simplified structs for receiving data (Input DTOs)
// This is used for POST/PUT requests when the client sends us data.
type CreateConfigurationRequest struct {
	Name    string      `json:"name"`
	Version string      `json:"version"`
	Params  []Parameter `json:"params"`
}

type CreateGroupRequest struct {
	Name           string          `json:"name"`
	Version        string          `json:"version"`
	Configurations []Configuration `json:"configurations"` // Using full config model for simplicity here
}
