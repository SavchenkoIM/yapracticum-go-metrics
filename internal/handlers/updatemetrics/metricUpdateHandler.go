package updatemetrics

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/handlers/middleware"
	"yaprakticum-go-track2/internal/storage"
	"yaprakticum-go-track2/internal/storage/storagecommons"

	"github.com/go-chi/chi/v5"
)

var dataStorage *storage.Storage

func SetDataStorage(storage *storage.Storage) {
	dataStorage = storage
}

var cfg config.ServerConfig

func SetCongig(config config.ServerConfig) {
	cfg = config
}

func checkHmacSha256(r *http.Request, cfg config.ServerConfig) error {

	if cfg.Key == "" {
		return nil
	}

	if r.Header.Get("HashSHA256") == "" {
		middleware.Logger.Info("No HashSHA256 header provided")
		return nil
	}

	hmacSha256, err := hex.DecodeString(r.Header.Get("HashSHA256"))
	if err != nil {
		middleware.Logger.Info("Incorrect Header HashSHA256")
		return err
	}

	b := make([]byte, r.ContentLength)
	_, err = r.Body.Read(b)
	if err != nil {
		middleware.Logger.Info("Error while reading BODY: " + err.Error())
		return err
	}
	err = r.Body.Close()
	if err != nil {
		middleware.Logger.Info("Error while closing BODY: " + err.Error())
		return err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(b))

	hmc := hmac.New(sha256.New, []byte(cfg.Key))
	hmc.Write(b)

	if !hmac.Equal(hmc.Sum(nil), hmacSha256) {
		middleware.Logger.Info("Incorrect HMAC SHA256")
		return errors.New("incorrect HMAC SHA256")
	}

	return nil
}

func MetricUpdateHandler(res http.ResponseWriter, req *http.Request) {

	typ := chi.URLParam(req, "type")
	name := chi.URLParam(req, "name")
	val := chi.URLParam(req, "value")

	switch typ {
	case "gauge":

		if err := dataStorage.GetGauges().WriteData(req.Context(), name, val); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

	case "counter":

		if err := dataStorage.GetCounters().WriteData(req.Context(), name, val); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

	default:

		http.Error(res, "Unknown metric type: "+typ, http.StatusBadRequest)
		return
	}

}

func MetricsUpdateHandlerREST(res http.ResponseWriter, req *http.Request) {

	if err := checkHmacSha256(req, cfg); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	var dta storagecommons.Metrics

	body := make([]byte, req.ContentLength)
	req.Body.Read(body)
	req.Body.Close()

	err := json.Unmarshal(body, &dta)
	if err != nil {
		http.Error(res, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	resp, err := dataStorage.WriteData(req.Context(), dta)
	val, _ := json.MarshalIndent(resp, "", "    ")

	if err == nil {
		res.Header().Set("Content-Type", "application/json")
		res.Write(val)
		return
	}

	http.Error(res, err.Error(), http.StatusBadRequest)
}

func MultiMetricsUpdateHandlerREST(res http.ResponseWriter, req *http.Request) {

	if err := checkHmacSha256(req, cfg); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	var dta storagecommons.MetricsDB

	body := make([]byte, req.ContentLength)
	req.Body.Read(body)
	req.Body.Close()

	err := json.Unmarshal(body, &dta.MetricsDB)
	if err != nil {
		http.Error(res, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	err = dataStorage.WriteDataMulty(req.Context(), dta)

	if err == nil {
		return
	}

	http.Error(res, err.Error(), http.StatusBadRequest)
}
