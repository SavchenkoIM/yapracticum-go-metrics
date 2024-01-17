package storage

import (
	_ "github.com/jackc/pgx/v5/stdlib"
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

func InitStorage(args config.ServerConfig, logger *zap.Logger) (*Storage, error) {
	var ms Storage

	println(args.ConnString)
	if args.ConnString == "" {
		fs, _ := filestore.New(args, logger)
		ms.Storager = fs
	} else {
		dbs, _ := dbstore.New(args, logger)
		ms.Storager = dbs
	}
	
	return &ms, nil
}
