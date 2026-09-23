package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Cunyule/AI-coding/internal/engine"
	"github.com/Cunyule/AI-coding/internal/model"
)

const maxRequestBytes = 8 << 20

func Handler(fingerprinter *engine.Engine) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /fingerprint", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)
		decoder := json.NewDecoder(r.Body)
		var inputs []model.ScanRecord
		if err := decoder.Decode(&inputs); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON request"})
			return
		}
		if err := ensureEOF(decoder); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "request must contain one JSON array"})
			return
		}
		writeJSON(w, http.StatusOK, fingerprinter.IdentifyBatch(inputs))
	})
	return mux
}

func ensureEOF(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if err == io.EOF {
		return nil
	}
	if err == nil {
		return fmt.Errorf("unexpected trailing JSON value")
	}
	return err
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
