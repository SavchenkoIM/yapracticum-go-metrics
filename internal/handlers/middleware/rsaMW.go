package middleware

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"hash"
	"io"
	"net/http"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/shared"
)

func DecryptOAEP(hash hash.Hash, random io.Reader, private *rsa.PrivateKey, msg []byte, label []byte) ([]byte, error) {
	msgLen := len(msg)
	step := private.PublicKey.Size()
	var decryptedBytes []byte

	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}

		decryptedBlockBytes, err := rsa.DecryptOAEP(hash, random, private, msg[start:finish], label)
		if err != nil {
			return nil, err
		}

		decryptedBytes = append(decryptedBytes, decryptedBlockBytes...)
	}

	return decryptedBytes, nil
}

func WithRSA(cfg config.ServerConfig) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.UseRSA {
				body, _ := io.ReadAll(r.Body)
				oaep, err := DecryptOAEP(sha256.New(), nil, &cfg.RSAPrivateKey, body, nil)
				if err != nil {
					shared.Logger.Error(err.Error())
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				r.Body = io.NopCloser(bytes.NewBuffer(oaep))
			}
			h.ServeHTTP(w, r)
		})
	}
}
