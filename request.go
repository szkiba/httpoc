package httpoc

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
)

func DecodeQuery(r *http.Request, dst interface{}) error {
	if err := decoder.Decode(dst, r.URL.Query()); err != nil {
		return ErrBadRequest.From(err)
	}

	if err := validate.Struct(dst); err != nil {
		return ErrBadRequest.From(err)
	}

	return nil
}

func DecodeForm(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return ErrBadRequest.From(err)
	}

	if err := decoder.Decode(dst, r.Form); err != nil {
		return ErrBadRequest.From(err)
	}

	if err := validate.Struct(dst); err != nil {
		return ErrBadRequest.From(err)
	}

	return nil
}

var (
	decoder  = schema.NewDecoder()
	validate = validator.New()
)
