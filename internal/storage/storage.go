package storage

import (
	"context"
	"go.uber.org/zap"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/storage/dbstore"
	"yaprakticum-go-track2/internal/storage/filestore"
	"yaprakticum-go-track2/internal/storage/storagecommons"
)

// Storage

type Storage struct {
	storagecommons.Storager
}

func InitStorage(ctx context.Context, args config.ServerConfig, logger *zap.Logger) (*Storage, error) {
	var ms Storage

	if args.ConnString == "" {
		fs, _ := filestore.New(ctx, args, logger)
		ms.Storager = fs
	} else {
		dbs, _ := dbstore.New(ctx, args, logger)
		ms.Storager = dbs
	}

	return &ms, nil
}
