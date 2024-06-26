package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/shared"
	"yaprakticum-go-track2/internal/storage/storagecommons"

	"github.com/go-chi/chi/v5"
)

// Auxilary method for checking HMAC signature of request
func checkHmacSha256(r *http.Request, cfg config.ServerConfig) error {

	if cfg.Key == "" {
		return nil
	}

	if r.Header.Get("HashSHA256") == "" {
		shared.Logger.Info("No HashSHA256 header provided")
		return nil
	}

	hmacSha256, err := hex.DecodeString(r.Header.Get("HashSHA256"))
	if err != nil {
		shared.Logger.Info("Incorrect Header HashSHA256")
		return err
	}

	b := make([]byte, r.ContentLength)
	_, err = r.Body.Read(b)
	if err != nil {
		shared.Logger.Info("Error while reading BODY: " + err.Error())
		return err
	}
	err = r.Body.Close()
	if err != nil {
		shared.Logger.Info("Error while closing BODY: " + err.Error())
		return err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(b))

	hmc := hmac.New(sha256.New, []byte(cfg.Key))
	hmc.Write(b)

	if !hmac.Equal(hmc.Sum(nil), hmacSha256) {
		shared.Logger.Info("Incorrect HMAC SHA256")
		return errors.New("incorrect HMAC SHA256")
	}

	return nil
}

// Storing metric data of given type and name
//
// Metric data is extracted from URL
func (h Handlers) MetricUpdateHandler(res http.ResponseWriter, req *http.Request) {

	typ := chi.URLParam(req, "type")
	name := chi.URLParam(req, "name")
	val := chi.URLParam(req, "value")

	switch typ {
	case "gauge":

		parsedVal, err := strconv.ParseFloat(val, 64)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		if err := h.dataStorage.WriteDataMulti(req.Context(), storagecommons.MetricsDB{
			MetricsDB: []storagecommons.Metrics{{Delta: nil, Value: &parsedVal, ID: name, MType: typ}}}); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		/*if err := h.dataStorage.GetGauges().WriteData(req.Context(), name, val); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}*/

	case "counter":

		parsedVal, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		if err := h.dataStorage.WriteDataMulti(req.Context(), storagecommons.MetricsDB{
			MetricsDB: []storagecommons.Metrics{{Delta: &parsedVal, Value: nil, ID: name, MType: typ}}}); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		/*if err := h.dataStorage.GetCounters().WriteData(req.Context(), name, val); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}*/

	default:

		http.Error(res, "Unknown metric type: "+typ, http.StatusBadRequest)
		return
	}

}

// Storing metric data of given type and name
//
// Metric data is expected to be JSON string and is extracted from request body
func (h Handlers) MetricsUpdateHandlerREST(res http.ResponseWriter, req *http.Request) {

	if err := checkHmacSha256(req, h.cfg); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	var dta storagecommons.Metrics

	body := make([]byte, req.ContentLength)
	req.Body.Read(body)
	req.Body.Close()

	err := json.Unmarshal(body, &dta)
	if err != nil {
		println(err.Error())
		http.Error(res, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	resp, err := h.dataStorage.WriteData(req.Context(), dta)
	val, _ := json.MarshalIndent(resp, "", "    ")

	if err == nil {
		res.Header().Set("Content-Type", "application/json")
		res.Write(val)
		return
	}

	http.Error(res, err.Error(), http.StatusBadRequest)
}

// Packet storing of metrics data
//
// Data is expected to be JSON string and is extracted from request body
func (h Handlers) MultiMetricsUpdateHandlerREST(res http.ResponseWriter, req *http.Request) {

	if err := checkHmacSha256(req, h.cfg); err != nil {
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

	err = h.dataStorage.WriteDataMulti(req.Context(), dta)

	if err == nil {
		return
	}

	http.Error(res, err.Error(), http.StatusBadRequest)
}
