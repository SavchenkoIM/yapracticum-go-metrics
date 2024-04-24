// Package contains tools for parsing Agent and Server runtime configuration data

package config

import (
	"crypto/rsa"
	"crypto/x509"
	"flag"
	"os"
	"strconv"
	"time"
)

// Server configuration
type ServerConfig struct {
	Endp            string
	EndpProm        string
	FileStoragePath string
	ConnString      string
	Key             string
	StoreInterval   time.Duration
	Restore         bool
	UseRSA          bool
	RSAPrivateKey   rsa.PrivateKey
}

// Parses Server configuration
func (cfg *ServerConfig) Load() ServerConfig {
	endp := flag.String("a", ":8080", "Server endpoint address:port")
	endpprom := flag.String("ap", ":18080", "Prom server endpoint address:port")
	storeInterval := flag.Int64("i", 300, "Store interval")
	fileStoragePath := flag.String("f", "/tmp/metrics-db.json", "File storage path")
	restoreData := flag.Bool("r", true, "Restore data from disc")
	connString := flag.String("d", "", "DB Connection string")
	key := flag.String("k", "", "Key")
	rsakey := flag.String("crypto-key", "", "RSA private key file name")
	flag.Parse()

	if val, exist := os.LookupEnv("ADDRESS"); exist {
		*endp = val
	}
	if val, exist := os.LookupEnv("ADDRESSPROM"); exist {
		*endpprom = val
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
	if val, exist := os.LookupEnv("DATABASE_DSN"); exist {
		*connString = val
	}
	if val, exist := os.LookupEnv("KEY"); exist {
		*key = val
	}
	if val, exist := os.LookupEnv("CRYPTO_KEY"); exist {
		*rsakey = val
	}

	cfg.Endp = *endp
	cfg.EndpProm = *endpprom
	cfg.FileStoragePath = *fileStoragePath
	cfg.Restore = *restoreData
	cfg.StoreInterval = time.Duration(*storeInterval) * time.Second
	cfg.ConnString = *connString
	cfg.Key = *key

	if *rsakey == "" {
		return *cfg
	}
	pk, err := os.ReadFile(*rsakey)
	if err != nil {
		cfg.UseRSA = false
		return *cfg
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(pk)
	if err != nil {
		cfg.UseRSA = false
		return *cfg
	}
	cfg.RSAPrivateKey = *privateKey
	cfg.UseRSA = true

	return *cfg
}

// Agent configuration
type ClientConfig struct {
	Endp           string
	Key            string
	ReqLimit       int64
	PollInterval   time.Duration
	ReportInterval time.Duration
	UseRSA         bool
	RSAPublicKey   rsa.PublicKey
}

// Parses Agent configuration
func (cfg *ClientConfig) Load() ClientConfig {
	endp := flag.String("a", "localhost:8080", "Server endpoint address:port")
	pollInterval := flag.Float64("p", 2, "pollInterval")
	reportInterval := flag.Float64("r", 10, "reportInterval")
	key := flag.String("k", "", "Key")
	rateLimit := flag.Int64("l", 5, "Limit of simultaneous requests")
	rsakey := flag.String("crypto-key", "", "RSA public key file name")
	flag.Parse()

	if val, exist := os.LookupEnv("ADDRESS"); exist {
		*endp = val
	}
	if _, exist := os.LookupEnv("REPORT_INTERVAL"); exist {
		if val, err := strconv.ParseFloat(os.Getenv("REPORT_INTERVAL"), 64); err != nil {
			*reportInterval = val
		}
	}
	if _, exist := os.LookupEnv("POLL_INTERVAL"); exist {
		if val, err := strconv.ParseFloat(os.Getenv("POLL_INTERVAL"), 64); err != nil {
			*pollInterval = val
		}
	}
	if val, exist := os.LookupEnv("KEY"); exist {
		*key = val
	}
	if _, exist := os.LookupEnv("RATE_LIMIT"); exist {
		if val, err := strconv.ParseInt(os.Getenv("POLL_INTERVAL"), 10, 0); err != nil {
			*rateLimit = val
		}
	}

	cfg.Endp = *endp
	cfg.PollInterval = time.Duration(*pollInterval) * time.Second
	cfg.ReportInterval = time.Duration(*reportInterval) * time.Second
	cfg.Key = *key
	cfg.ReqLimit = *rateLimit

	if *rsakey == "" {
		return *cfg
	}
	pk, err := os.ReadFile(*rsakey)
	if err != nil {
		println(err.Error())
		cfg.UseRSA = false
		return *cfg
	}
	publicKey, err := x509.ParsePKCS1PublicKey(pk)
	if err != nil {
		println(err.Error())
		cfg.UseRSA = false
		return *cfg
	}
	cfg.RSAPublicKey = *publicKey
	cfg.UseRSA = true

	return *cfg
}
