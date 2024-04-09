// Package contains handlers of metrics and alerting server

package handlers

import (
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/storage"
)

// Handlers object of metrics and alerting server
type Handlers struct {
	dataStorage *storage.Storage
	cfg         config.ServerConfig
}

// Constructor of Handlers
func NewHandlers(storage *storage.Storage, config config.ServerConfig) Handlers {
	return Handlers{dataStorage: storage, cfg: config}
}
