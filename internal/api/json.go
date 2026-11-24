package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/pablogtrz/tablon/internal/domain"
)

const maxBodyBytes = 1 << 20

type nullableString struct {
	Set   bool
	Value *string
}

func (n *nullableString) UnmarshalJSON(data []byte) error {
	n.Set = true
	if string(data) == "null" {
		n.Value = nil
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	n.Value = &value
	return nil
}

type errorResponse struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if status == http.StatusNoContent {
		return
	}
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code, message string, details any) {
	writeJSON(w, status, errorResponse{Error: errorDetail{
		Code:    code,
		Message: message,
		Details: details,
	}})
}

func handleError(w http.ResponseWriter, logger *slog.Logger, err error) {
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "resource not found", nil)
		return
	}
	if errors.Is(err, domain.ErrConflict) {
		writeError(w, http.StatusConflict, "conflict", "resource state conflicts with the request", nil)
		return
	}
	if ve, ok := domain.AsValidation(err); ok {
		details := map[string]string{"field": ve.Field}
		writeError(w, http.StatusUnprocessableEntity, "validation_error", ve.Error(), details)
		return
	}
	logger.Error("request failed", "error", err)
	writeError(w, http.StatusInternalServerError, "internal_error", "internal server error", nil)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	_, err := decodeJSONWithBody(w, r, dst)
	return err
}

// decodeJSONWithBody decodes the request and also returns the raw body so
// callers can fingerprint it for Idempotency-Key handling.
func decodeJSONWithBody(w http.ResponseWriter, r *http.Request, dst any) ([]byte, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return nil, err
	}
	var extra struct{}
	if err := dec.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("body must contain a single JSON object")
	}
	return raw, nil
}

func writeDecodeError(w http.ResponseWriter, err error) {
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) {
		writeError(w, http.StatusRequestEntityTooLarge, "request_too_large", "request body is too large", nil)
		return
	}
	writeError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
}
