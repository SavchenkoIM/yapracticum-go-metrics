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

// Agent configuration
type ClientConfig struct {
	Endp           string
	Key            string
	ReqLimit       int64
	PollInterval   time.Duration
	ReportInterval time.Duration
	UseRSA         bool
	RSAPublicKey   rsa.PublicKey
	RealIP         net.IP
}

type clientConfigNull struct {
	Endp             *string
	Key              *string
	ReqLimit         *int64
	PollInterval     *time.Duration
	ReportInterval   *time.Duration
	UseRSA           *bool
	RSAPublicKeyFile *string
	ConfigFile       *string
}

type ClientConfigFile struct {
	Address        *string `json:"address,omitempty"`
	ReportInterval *string `json:"report_interval,omitempty"`
	PollInterval   *string `json:"poll_interval,omitempty"`
	CryptoKey      *string `json:"crypto_key,omitempty"`
}

// Parses Agent configuration from Command Line args
func getClientConfigFromCLArgs() clientConfigNull {
	clientConfig := clientConfigNull{}
	endp := flag.String("a", "localhost:8080", "Server endpoint address:port")
	pollInterval := flag.Float64("p", 2, "pollInterval")
	reportInterval := flag.Float64("r", 10, "reportInterval")
	key := flag.String("k", "", "Key")
	rateLimit := flag.Int64("l", 5, "Limit of simultaneous requests")
	rsakey := flag.String("crypto-key", "", "RSA public key file name")
	configFile := flag.String("c", "", "Config file")
	flag.StringVar(configFile, "config", "", "Config file")
	flag.Parse()

	usedFlags := getProvidedFlags(flag.Visit)

	clientConfig.Endp = getParWithSetCheck[string](*endp, slices.Contains(usedFlags, "a"))
	clientConfig.PollInterval = getParWithSetCheck[time.Duration](time.Duration(*pollInterval)*time.Second, slices.Contains(usedFlags, "p"))
	clientConfig.ReportInterval = getParWithSetCheck[time.Duration](time.Duration(*reportInterval)*time.Second, slices.Contains(usedFlags, "r"))
	clientConfig.Key = getParWithSetCheck[string](*key, slices.Contains(usedFlags, "k"))
	clientConfig.ReqLimit = getParWithSetCheck[int64](*rateLimit, slices.Contains(usedFlags, "l"))
	clientConfig.RSAPublicKeyFile = getParWithSetCheck[string](*rsakey, slices.Contains(usedFlags, "crypto-key") || slices.Contains(usedFlags, "c"))
	clientConfig.ConfigFile = getParWithSetCheck[string](*configFile, slices.Contains(usedFlags, "c") || slices.Contains(usedFlags, "config"))

	return clientConfig
}

// Parses Agent configuration from Enviroment Vars
func getClientConfigFromEnvVar() clientConfigNull {
	clientConfig := clientConfigNull{}
	endp := envflag.String("ADDRESS", ":8080", "Server endpoint address:port")
	pollInterval := envflag.Float64("POLL_INTERVAL", 2, "pollInterval")
	reportInterval := envflag.Float64("REPORT_INTERVAL", 10, "reportInterval")
	key := envflag.String("KEY", "", "Key")
	rateLimit := envflag.Int64("RATE_LIMIT", 5, "Limit of simultaneous requests")
	rsakey := envflag.String("CRYPTO_KEY", "", "RSA public key file name")
	configFile := envflag.String("CONFIG", "", "Config file")
	envflag.Parse()

	usedFlags := getProvidedFlags(envflag.Visit)

	clientConfig.Endp = getParWithSetCheck[string](*endp, slices.Contains(usedFlags, "ADDRESS"))
	clientConfig.PollInterval = getParWithSetCheck[time.Duration](time.Duration(*pollInterval)*time.Second, slices.Contains(usedFlags, "POLL_INTERVAL"))
	clientConfig.ReportInterval = getParWithSetCheck[time.Duration](time.Duration(*reportInterval)*time.Second, slices.Contains(usedFlags, "REPORT_INTERVAL"))
	clientConfig.Key = getParWithSetCheck[string](*key, slices.Contains(usedFlags, "KEY"))
	clientConfig.ReqLimit = getParWithSetCheck[int64](*rateLimit, slices.Contains(usedFlags, "RATE_LIMIT"))
	clientConfig.RSAPublicKeyFile = getParWithSetCheck[string](*rsakey, slices.Contains(usedFlags, "CRYPTO_KEY"))
	clientConfig.ConfigFile = getParWithSetCheck[string](*configFile, slices.Contains(usedFlags, "CONFIG"))

	return clientConfig
}

// Parses Server configuration from ConfigFile
func getClientConfigFromJSON(filename string) clientConfigNull {
	clientConfig := clientConfigNull{}
	if filename == "" {
		return clientConfig
	}

	ccf := ClientConfigFile{}
	cfile, err := os.ReadFile(filename)
	if err != nil {
		return clientConfig
	}

	err = json.Unmarshal(cfile, &ccf)
	if err != nil {
		return clientConfig
	}

	clientConfig.Endp = ccf.Address
	clientConfig.PollInterval = getDurationFromString(ccf.PollInterval)
	clientConfig.ReportInterval = getDurationFromString(ccf.ReportInterval)
	clientConfig.Key = nil
	clientConfig.ReqLimit = nil
	clientConfig.RSAPublicKeyFile = ccf.CryptoKey
	clientConfig.ConfigFile = nil

	return clientConfig
}

func CombineClientConfigs(configs ...clientConfigNull) ClientConfig {
	clientConfig := ClientConfig{
		Endp:           ":8080",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		ReqLimit:       5,
		Key:            "",
		UseRSA:         false,
		RSAPublicKey:   rsa.PublicKey{},
	}

	slices.Reverse(configs)
	for _, cfg := range configs {
		combineParameter(&clientConfig.Endp, cfg.Endp)
		combineParameter(&clientConfig.PollInterval, cfg.PollInterval)
		combineParameter(&clientConfig.ReportInterval, cfg.ReportInterval)
		combineParameter(&clientConfig.ReqLimit, cfg.ReqLimit)
		combineParameter(&clientConfig.Key, cfg.Key)
		var (
			rsaUse bool
			rsaKey rsa.PublicKey
		)
		if cfg.RSAPublicKeyFile != nil {
			rsaKey, rsaUse = getRSAPublicKey(*cfg.RSAPublicKeyFile)
		}
		if rsaUse {
			clientConfig.RSAPublicKey = rsaKey
			clientConfig.UseRSA = rsaUse
		}
	}

	return clientConfig
}

// Parses Agent configuration
func (cfg *ClientConfig) Load() ClientConfig {

	clConf := getClientConfigFromCLArgs()
	envConf := getClientConfigFromEnvVar()

	confFileName := ""
	if clConf.ConfigFile != nil {
		confFileName = *clConf.ConfigFile
	}
	if envConf.ConfigFile != nil {
		confFileName = *clConf.ConfigFile
	}

	fileConf := getClientConfigFromJSON(confFileName)

	*cfg = CombineClientConfigs(envConf, clConf, fileConf)
	cfg.RealIP = getPreferredIP(cfg.Endp)

	return *cfg
}
