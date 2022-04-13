package httpoc

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Failure struct {
	Code        string `json:"error,omitempty" schema:"error,omitempty"`
	Status      int    `json:"status" schema:"status"`
	Description string `json:"error_description,omitempty" schema:"error_description,omitempty"`
}

func (f *Failure) Error() string {
	return f.Code
}

func (f *Failure) From(err error) *Failure {
	return &Failure{
		Code:        f.Code,
		Status:      f.Status,
		Description: err.Error(),
	}
}

func (f *Failure) MarshalZerologObject(e *zerolog.Event) {
	e.Str("error", f.Code)

	if len(f.Description) != 0 {
		e.Str("error_description", f.Description)
	}
}

func toFailure(r *http.Request, err error) *Failure {
	var f *Failure
	if !errors.As(err, &f) {
		f = ErrServerError.From(err)
	}

	if m, ok := r.Context().Value(metricsKey{}).(*metrics); ok {
		m.Failure = f
	}

	return f
}

func WriteFailure(w http.ResponseWriter, r *http.Request, err error) {
	f := toFailure(r, err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(f.Status)

	if err := json.NewEncoder(w).Encode(f); err != nil {
		log.Warn().Err(err).Msg("failed to send error response")
	}
}

var (
	ErrUnauthorized = &Failure{Code: "unauthorized", Status: http.StatusUnauthorized}
	ErrServerError  = &Failure{Code: "server_error", Status: http.StatusInternalServerError}
	ErrBadRequest   = &Failure{Code: "invalid_request", Status: http.StatusBadRequest}
	ErrNotFound     = &Failure{Code: "not_found", Status: http.StatusNotFound}
)
