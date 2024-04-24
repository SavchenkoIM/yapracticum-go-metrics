package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"os"
)

func main() {
	fileName := `cert\metrics`

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		println(err.Error())
	}

	err = os.WriteFile(fileName+".key", x509.MarshalPKCS1PrivateKey(key), 0644)
	if err != nil {
		println(err.Error())
	}

	err = os.WriteFile(fileName+"_pub.key", x509.MarshalPKCS1PublicKey(&key.PublicKey), 0644)
	if err != nil {
		println(err.Error())
	}
}
