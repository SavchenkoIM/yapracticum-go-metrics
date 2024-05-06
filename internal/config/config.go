// Package contains tools for parsing Agent and Server runtime configuration data

package config

import (
	"crypto/rsa"
	"crypto/x509"
	"flag"
	"os"
	"reflect"
	"time"
)

const (
	ParType_String = iota
	ParType_Int64
	ParType_Duration
	PrType_Bool
)

type configParam struct {
	ParType reflect.Type
	ParName string
	IsSet   bool
	Value   interface{}
}

func getProvidedFlags(visitFunc func(func(f *flag.Flag))) []string {
	res := make([]string, 0)
	visitFunc(func(f *flag.Flag) {
		res = append(res, f.Name)
	})
	return res
}

// Returns duraion from string representation (of nil if failed)
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

func getParWithSetCheck[S any](val S, isSet bool) *S {
	if !isSet {
		return nil
	}
	return &val
}

func combineParameter[S any](dst *S, src *S) {
	if src == nil {
		return
	}
	*dst = *src
}

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
