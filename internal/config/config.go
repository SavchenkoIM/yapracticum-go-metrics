package config

import (
	"flag"
	"os"
	"strconv"
	"time"
)

type ServerConfig struct {
	Endp            string
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
}

func (cfg *ServerConfig) Load() ServerConfig {
	endp := flag.String("a", ":8080", "Server endpoint address:port")
	storeInterval := flag.Int64("i", 300, "Store interval")
	fileStoragePath := flag.String("f", "/tmp/metrics-db.json", "File storage path")
	restoreData := flag.Bool("r", true, "Restore data from disc")
	flag.Parse()

	if val, exist := os.LookupEnv("ADDRESS"); exist {
		*endp = val
	}
	if _, exist := os.LookupEnv("STORE_INTERVAL"); exist {
		if val, err := strconv.ParseInt(os.Getenv("STORE_INTERVAL"), 10, 64); err != nil {
			*storeInterval = val
		}
	}
	if val, exist := os.LookupEnv("FILE_STORAGE_PATH"); exist {
		*fileStoragePath = val
	}
	if _, exist := os.LookupEnv("RESTORE"); exist {
		if val, err := strconv.ParseBool(os.Getenv("RESTORE")); err != nil {
			*restoreData = val
		}
	}

	cfg.Endp = *endp
	cfg.FileStoragePath = *fileStoragePath
	cfg.Restore = *restoreData
	cfg.StoreInterval = time.Duration(*storeInterval) * time.Second

	return *cfg
}
