// Package contains handlers of metrics and alerting server

package handlers

import (
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/storage"
)

// "Storage" of the package
var dataStorage *storage.Storage

// Sets object "storage" for the package
func SetDataStorage(storage *storage.Storage) {
	dataStorage = storage
}

// Server configuration parameters
var cfg config.ServerConfig

// Sets server configuration parameters for the package
func SetConfig(config config.ServerConfig) {
	cfg = config
}
