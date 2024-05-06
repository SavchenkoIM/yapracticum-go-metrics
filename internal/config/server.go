package config

import (
	"crypto/rsa"
	"encoding/json"
	"flag"
	"github.com/ianschenck/envflag"
	"os"
	"slices"
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
	TrustedSubnet   string
	UseRSA          bool
	RSAPrivateKey   rsa.PrivateKey
}

// Raw server configuration with possible null fields
type serverConfigNull struct {
	Endp              *string
	EndpProm          *string
	FileStoragePath   *string
	ConnString        *string
	Key               *string
	StoreInterval     *time.Duration
	Restore           *bool
	TrustedSubnet     *string
	RSAPrivateKeyFile *string
	ConfigFile        *string
}

// Representation of JSON config file
type ServerConfigFile struct {
	Address       *string `json:"address,omitempty"`
	Restore       *bool   `json:"restore,omitempty"`
	StoreInterval *string `json:"store_interval,omitempty"`
	StoreFile     *string `json:"store_file,omitempty"`
	DatabaseDsn   *string `json:"database_dsn,omitempty"`
	TrustedSubnet *string `json:"trusted_subnet,omitempty"`
	CryptoKey     *string `json:"crypto_key,omitempty"`
}

// Parses Server configuration from Command Line args
func getServerConfigFromCLArgs() serverConfigNull {
	serverConfig := serverConfigNull{}
	endp := flag.String("a", ":8080", "Server endpoint address:port")
	endpprom := flag.String("ap", ":18080", "Prom server endpoint address:port")
	storeInterval := flag.Int64("i", 300, "Store interval")
	fileStoragePath := flag.String("f", "/tmp/metrics-db.json", "File storage path")
	restoreData := flag.Bool("r", true, "Restore data from disc")
	connString := flag.String("d", "", "DB Connection string")
	key := flag.String("k", "", "Key")
	trustedSubnet := flag.String("t", "", "Trusted Subnet")
	rsakey := flag.String("crypto-key", "", "RSA private key file name")
	configFile := flag.String("c", "", "Config file")
	flag.StringVar(configFile, "config", "", "Config file")
	flag.Parse()

	usedFlags := getProvidedFlags(flag.Visit)

	serverConfig.Endp = getParWithSetCheck[string](*endp, slices.Contains(usedFlags, "a"))
	serverConfig.EndpProm = getParWithSetCheck[string](*endpprom, slices.Contains(usedFlags, "ap"))
	serverConfig.StoreInterval = getParWithSetCheck[time.Duration](time.Duration(*storeInterval)*time.Second, slices.Contains(usedFlags, "i"))
	serverConfig.FileStoragePath = getParWithSetCheck[string](*fileStoragePath, slices.Contains(usedFlags, "f"))
	serverConfig.Restore = getParWithSetCheck[bool](*restoreData, slices.Contains(usedFlags, "r"))
	serverConfig.TrustedSubnet = getParWithSetCheck[string](*trustedSubnet, slices.Contains(usedFlags, "t"))
	serverConfig.ConnString = getParWithSetCheck[string](*connString, slices.Contains(usedFlags, "d"))
	serverConfig.Key = getParWithSetCheck[string](*key, slices.Contains(usedFlags, "k"))
	serverConfig.RSAPrivateKeyFile = getParWithSetCheck[string](*rsakey, slices.Contains(usedFlags, "crypto-key") || slices.Contains(usedFlags, "c"))
	serverConfig.ConfigFile = getParWithSetCheck[string](*configFile, slices.Contains(usedFlags, "c") || slices.Contains(usedFlags, "config"))

	return serverConfig
}

// Parses Server configuration from Enviroment Vars
func getServerConfigFromEnvVar() serverConfigNull {
	serverConfig := serverConfigNull{}
	endp := envflag.String("ADDRESS", ":8080", "Server endpoint address:port")
	endpprom := envflag.String("ADDRESSPROM", ":18080", "Prom server endpoint address:port")
	storeInterval := envflag.Int64("STORE_INTERVAL", 300, "Store interval")
	fileStoragePath := envflag.String("FILE_STORAGE_PATH", "/tmp/metrics-db.json", "File storage path")
	restoreData := envflag.Bool("RESTORE", true, "Restore data from disc")
	connString := envflag.String("DATABASE_DSN", "", "DB Connection string")
	key := envflag.String("KEY", "", "Key")
	trustedSubnet := flag.String("TRUSTED_SUBNET", "", "Trusted Subnet")
	rsakey := envflag.String("CRYPTO_KEY", "", "RSA private key file name")
	configFile := envflag.String("CONFIG", "", "Config file")
	envflag.Parse()

	usedFlags := getProvidedFlags(envflag.Visit)

	serverConfig.Endp = getParWithSetCheck[string](*endp, slices.Contains(usedFlags, "ADDRESS"))
	serverConfig.EndpProm = getParWithSetCheck[string](*endpprom, slices.Contains(usedFlags, "ADDRESSPROM"))
	serverConfig.StoreInterval = getParWithSetCheck[time.Duration](time.Duration(*storeInterval)*time.Second, slices.Contains(usedFlags, "STORE_INTERVAL"))
	serverConfig.FileStoragePath = getParWithSetCheck[string](*fileStoragePath, slices.Contains(usedFlags, "FILE_STORAGE_PATH"))
	serverConfig.Restore = getParWithSetCheck[bool](*restoreData, slices.Contains(usedFlags, "RESTORE"))
	serverConfig.TrustedSubnet = getParWithSetCheck(*trustedSubnet, slices.Contains(usedFlags, "TRUSTED_SUBNET"))
	serverConfig.ConnString = getParWithSetCheck[string](*connString, slices.Contains(usedFlags, "DATABASE_DSN"))
	serverConfig.Key = getParWithSetCheck[string](*key, slices.Contains(usedFlags, "KEY"))
	serverConfig.RSAPrivateKeyFile = getParWithSetCheck[string](*rsakey, slices.Contains(usedFlags, "CRYPTO_KEY"))
	serverConfig.ConfigFile = getParWithSetCheck[string](*configFile, slices.Contains(usedFlags, "CONFIG"))

	return serverConfig
}

// Parses Server configuration from ConfigFile
func getServerConfigFromJSON(filename string) serverConfigNull {
	serverConfig := serverConfigNull{}
	if filename == "" {
		return serverConfig
	}

	scf := ServerConfigFile{}
	cfile, err := os.ReadFile(filename)
	if err != nil {
		return serverConfig
	}

	err = json.Unmarshal(cfile, &scf)
	if err != nil {
		return serverConfig
	}

	serverConfig.Endp = scf.Address
	serverConfig.EndpProm = nil
	serverConfig.StoreInterval = getDurationFromString(scf.StoreInterval)
	serverConfig.FileStoragePath = scf.StoreFile
	serverConfig.Restore = scf.Restore
	serverConfig.TrustedSubnet = scf.TrustedSubnet
	serverConfig.ConnString = scf.DatabaseDsn
	serverConfig.Key = nil
	serverConfig.RSAPrivateKeyFile = scf.CryptoKey

	return serverConfig
}

func CombineServerConfigs(configs ...serverConfigNull) ServerConfig {
	serverConfig := ServerConfig{
		Endp:            ":8080",
		EndpProm:        ":18080",
		FileStoragePath: "/tmp/metrics-db.json",
		ConnString:      "",
		Key:             "",
		StoreInterval:   300,
		Restore:         true,
		UseRSA:          false,
		RSAPrivateKey:   rsa.PrivateKey{},
	}

	slices.Reverse(configs)
	for _, cfg := range configs {
		combineParameter(&serverConfig.Endp, cfg.Endp)
		combineParameter(&serverConfig.EndpProm, cfg.EndpProm)
		combineParameter(&serverConfig.FileStoragePath, cfg.FileStoragePath)
		combineParameter(&serverConfig.ConnString, cfg.ConnString)
		combineParameter(&serverConfig.Key, cfg.Key)
		combineParameter(&serverConfig.StoreInterval, cfg.StoreInterval)
		combineParameter(&serverConfig.Restore, cfg.Restore)
		combineParameter(&serverConfig.TrustedSubnet, cfg.TrustedSubnet)
		var (
			rsaUse bool
			rsaKey rsa.PrivateKey
		)
		if cfg.RSAPrivateKeyFile != nil {
			rsaKey, rsaUse = getRSAPrivateKey(*cfg.RSAPrivateKeyFile)
		}
		if rsaUse {
			serverConfig.RSAPrivateKey = rsaKey
			serverConfig.UseRSA = rsaUse
		}
	}

	return serverConfig
}

// Parses Server configuration
func (cfg *ServerConfig) Load() ServerConfig {

	clConf := getServerConfigFromCLArgs()
	envConf := getServerConfigFromEnvVar()

	confFileName := ""
	if clConf.ConfigFile != nil {
		confFileName = *clConf.ConfigFile
	}
	if envConf.ConfigFile != nil {
		confFileName = *clConf.ConfigFile
	}

	fileConf := getServerConfigFromJSON(confFileName)

	return CombineServerConfigs(envConf, clConf, fileConf)
}
