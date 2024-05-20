package config

import (
	"crypto/rsa"
	"encoding/json"
	"flag"
	"github.com/ianschenck/envflag"
	"net"
	"os"
	"slices"
	"time"
)

// Server configuration
type ServerConfig struct {
	Endp                string
	EndpProm            string
	EndpGRPC            string
	FileStoragePath     string
	ConnString          string
	Key                 string
	StoreInterval       time.Duration
	Restore             bool
	TrustedSubnet       *net.IPNet
	UseRSA              bool
	RSAPrivateKey       rsa.PrivateKey
	BandwidthPriority   bool
	CachedWriteInterval time.Duration
}

// Raw server configuration with possible null fields
type serverConfigNull struct {
	Endp                *string
	EndpProm            *string
	EndpGRPC            *string
	FileStoragePath     *string
	ConnString          *string
	Key                 *string
	StoreInterval       *time.Duration
	Restore             *bool
	TrustedSubnet       *string
	RSAPrivateKeyFile   *string
	BandwidthPriority   *bool
	CachedWriteInterval *time.Duration
	ConfigFile          *string
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
	endpgrpc := flag.String("ag", ":3200", "gRPC server endpoint address:port")
	storeInterval := flag.Int64("i", 300, "Store interval")
	fileStoragePath := flag.String("f", "/tmp/metrics-db.json", "File storage path")
	restoreData := flag.Bool("r", true, "Restore data from disc")
	connString := flag.String("d", "", "DB Connection string")
	key := flag.String("k", "", "Key")
	trustedSubnet := flag.String("t", "", "Trusted Subnet")
	rsakey := flag.String("crypto-key", "", "RSA private key file name")
	cachedWriteInterval := flag.Int64("cwi", 0, "Cached write interval, ms")
	configFile := flag.String("c", "", "Config file")
	flag.StringVar(configFile, "config", "", "Config file")
	flag.Parse()

	usedFlags := getProvidedFlags(flag.Visit)

	serverConfig.Endp = getParWithSetCheck(*endp, slices.Contains(usedFlags, "a"))
	serverConfig.EndpProm = getParWithSetCheck(*endpprom, slices.Contains(usedFlags, "ap"))
	serverConfig.EndpGRPC = getParWithSetCheck(*endpgrpc, slices.Contains(usedFlags, "ag"))
	serverConfig.StoreInterval = getParWithSetCheck(time.Duration(*storeInterval)*time.Second, slices.Contains(usedFlags, "i"))
	serverConfig.FileStoragePath = getParWithSetCheck(*fileStoragePath, slices.Contains(usedFlags, "f"))
	serverConfig.Restore = getParWithSetCheck(*restoreData, slices.Contains(usedFlags, "r"))
	serverConfig.TrustedSubnet = getParWithSetCheck(*trustedSubnet, slices.Contains(usedFlags, "t"))
	serverConfig.ConnString = getParWithSetCheck(*connString, slices.Contains(usedFlags, "d"))
	serverConfig.Key = getParWithSetCheck(*key, slices.Contains(usedFlags, "k"))
	serverConfig.RSAPrivateKeyFile = getParWithSetCheck(*rsakey, slices.Contains(usedFlags, "crypto-key") || slices.Contains(usedFlags, "c"))
	serverConfig.CachedWriteInterval = getParWithSetCheck(time.Duration(*cachedWriteInterval)*time.Millisecond, slices.Contains(usedFlags, "cwi"))
	serverConfig.ConfigFile = getParWithSetCheck(*configFile, slices.Contains(usedFlags, "c") || slices.Contains(usedFlags, "config"))

	return serverConfig
}

// Parses Server configuration from Enviroment Vars
func getServerConfigFromEnvVar() serverConfigNull {
	serverConfig := serverConfigNull{}
	endp := envflag.String("ADDRESS", ":8080", "Server endpoint address:port")
	endpprom := envflag.String("ADDRESS_PROM", ":18080", "Prom server endpoint address:port")
	endpgrpc := flag.String("ADDRESS_GRPC", ":3200", "gRPC server endpoint address:port")
	storeInterval := envflag.Int64("STORE_INTERVAL", 300, "Store interval")
	fileStoragePath := envflag.String("FILE_STORAGE_PATH", "/tmp/metrics-db.json", "File storage path")
	restoreData := envflag.Bool("RESTORE", true, "Restore data from disc")
	connString := envflag.String("DATABASE_DSN", "", "DB Connection string")
	key := envflag.String("KEY", "", "Key")
	trustedSubnet := envflag.String("TRUSTED_SUBNET", "", "Trusted Subnet")
	rsakey := envflag.String("CRYPTO_KEY", "", "RSA private key file name")
	cachedWriteInterval := envflag.Int64("CACHED_WRITE_INTERVAL", 0, "Cached write interval, ms")
	configFile := envflag.String("CONFIG", "", "Config file")
	envflag.Parse()

	usedFlags := getProvidedFlags(envflag.Visit)

	serverConfig.Endp = getParWithSetCheck(*endp, slices.Contains(usedFlags, "ADDRESS"))
	serverConfig.EndpProm = getParWithSetCheck(*endpprom, slices.Contains(usedFlags, "ADDRESS_PROM"))
	serverConfig.EndpGRPC = getParWithSetCheck(*endpgrpc, slices.Contains(usedFlags, "ag"))
	serverConfig.StoreInterval = getParWithSetCheck(time.Duration(*storeInterval)*time.Second, slices.Contains(usedFlags, "STORE_INTERVAL"))
	serverConfig.FileStoragePath = getParWithSetCheck(*fileStoragePath, slices.Contains(usedFlags, "FILE_STORAGE_PATH"))
	serverConfig.Restore = getParWithSetCheck(*restoreData, slices.Contains(usedFlags, "RESTORE"))
	serverConfig.TrustedSubnet = getParWithSetCheck(*trustedSubnet, slices.Contains(usedFlags, "TRUSTED_SUBNET"))
	serverConfig.ConnString = getParWithSetCheck(*connString, slices.Contains(usedFlags, "DATABASE_DSN"))
	serverConfig.Key = getParWithSetCheck(*key, slices.Contains(usedFlags, "KEY"))
	serverConfig.RSAPrivateKeyFile = getParWithSetCheck(*rsakey, slices.Contains(usedFlags, "CRYPTO_KEY"))
	serverConfig.CachedWriteInterval = getParWithSetCheck(time.Duration(*cachedWriteInterval)*time.Millisecond, slices.Contains(usedFlags, "cwi"))
	serverConfig.ConfigFile = getParWithSetCheck(*configFile, slices.Contains(usedFlags, "CONFIG"))

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
	serverConfig.CachedWriteInterval = nil

	return serverConfig
}

func CombineServerConfigs(configs ...serverConfigNull) ServerConfig {
	serverConfig := ServerConfig{
		Endp:              ":8080",
		EndpProm:          "", // 18080
		EndpGRPC:          "", // 3200
		FileStoragePath:   "/tmp/metrics-db.json",
		ConnString:        "",
		Key:               "",
		StoreInterval:     300,
		Restore:           true,
		UseRSA:            false,
		RSAPrivateKey:     rsa.PrivateKey{},
		TrustedSubnet:     nil,
		BandwidthPriority: false,
	}

	slices.Reverse(configs)
	for _, cfg := range configs {
		combineParameter(&serverConfig.Endp, cfg.Endp)
		combineParameter(&serverConfig.EndpProm, cfg.EndpProm)
		combineParameter(&serverConfig.EndpGRPC, cfg.EndpGRPC)
		combineParameter(&serverConfig.FileStoragePath, cfg.FileStoragePath)
		combineParameter(&serverConfig.ConnString, cfg.ConnString)
		combineParameter(&serverConfig.Key, cfg.Key)
		combineParameter(&serverConfig.StoreInterval, cfg.StoreInterval)
		combineParameter(&serverConfig.Restore, cfg.Restore)
		combineParameter(&serverConfig.CachedWriteInterval, cfg.CachedWriteInterval)

		// Caching
		if cfg.CachedWriteInterval != nil && *cfg.CachedWriteInterval > time.Duration(0) {
			serverConfig.BandwidthPriority = true
		}

		// Trusted subnet
		if cfg.TrustedSubnet != nil {
			_, ipNet, err := net.ParseCIDR(*cfg.TrustedSubnet)
			if err == nil {
				serverConfig.TrustedSubnet = ipNet
			}
		}

		// RSA
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

	*cfg = CombineServerConfigs(envConf, clConf, fileConf)
	return *cfg
}
