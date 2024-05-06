// Package contains tools for parsing Agent and Server runtime configuration data

package config

import (
	"crypto/rsa"
	"crypto/x509"
	"flag"
	"net"
	"os"
	"time"
)

// Gets provided Command Line flags or and Enviroment Vars of configuration
func getProvidedFlags(visitFunc func(func(f *flag.Flag))) []string {
	res := make([]string, 0)
	visitFunc(func(f *flag.Flag) {
		res = append(res, f.Name)
	})
	return res
}

// Returns duration from string representation (of nil if failed)
func getDurationFromString(sRepr *string) *time.Duration {
	if sRepr == nil {
		return nil
	}
	d, err := time.ParseDuration(*sRepr)
	if err != nil {
		return nil
	}
	return &d
}

// Returns nil if parameter does not set, otherwise pointer to the parameter
func getParWithSetCheck[S any](val S, isSet bool) *S {
	if !isSet {
		return nil
	}
	return &val
}

// Sets dst value equal to src value if src is not nil, otherwise do nothing
func combineParameter[S any](dst *S, src *S) {
	if src == nil {
		return
	}
	*dst = *src
}

// Returns RSA Private Key object stored in file
func getRSAPrivateKey(filename string) (PK rsa.PrivateKey, UseRSA bool) {
	var key rsa.PrivateKey
	if filename == "" {
		return key, false
	}
	pk, err := os.ReadFile(filename)
	if err != nil {
		return key, false
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(pk)
	if err != nil {
		return key, false
	}
	return *privateKey, true
}

// Returns RSA Public Key object stored in file
func getRSAPublicKey(filename string) (PK rsa.PublicKey, UseRSA bool) {
	var key rsa.PublicKey
	if filename == "" {
		return key, false
	}
	pk, err := os.ReadFile(filename)
	if err != nil {
		return key, false
	}
	publicKey, err := x509.ParsePKCS1PublicKey(pk)
	if err != nil {
		return key, false
	}
	return *publicKey, true
}

// Returns IP of interface, used to connect to desired IP
func getPreferredIP(url string) net.IP {
	conn, err := net.Dial("udp", url)
	if err != nil {
		return nil
	}
	return conn.LocalAddr().(*net.UDPAddr).IP
}
