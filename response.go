package httpoc

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

func WriteJSON(w http.ResponseWriter, resp interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Warn().Err(err).Msg("failed to send error response")
	}
}

type Result struct {
	Data interface{}
	Err  error
}

func Wrap(data interface{}, err error) *Result {
	return &Result{data, err}
}

func WriteResult(w http.ResponseWriter, r *http.Request, result *Result) {
	if result.Err == nil {
		WriteJSON(w, result.Data)
	} else {
		WriteFailure(w, r, result.Err)
	}
}
