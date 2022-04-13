package httpoc

import (
	"context"
	"net/http"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/rs/zerolog"
)

type metrics struct {
	Duration time.Duration
	Status   int
	Route    string
	Path     string
	Failure  *Failure
}

type metricsKey struct{}

func (m *metrics) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, metricsKey{}, m)
}

func (m *metrics) MarshalZerologObject(e *zerolog.Event) {
	e.Str("route", m.Route)
	e.Int("status", m.Status)
	e.Dur("duration", m.Duration)

	if m.Path != "" {
		e.Str("path", m.Path)
	}
}

func (m *metrics) logLevel() zerolog.Level {
	if m.Status >= http.StatusInternalServerError {
		return zerolog.ErrorLevel
	}

	if m.Status >= http.StatusBadRequest {
		return zerolog.WarnLevel
	}

	if m.Route == "/" {
		return zerolog.DebugLevel
	}

	return zerolog.InfoLevel
}

type actionHandler struct {
	route string
	next  http.Handler
}

func (h *actionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := zerolog.DefaultContextLogger.Level(zerolog.DefaultContextLogger.GetLevel())
	r = r.WithContext(l.WithContext(r.Context()))

	t := h.trace(w, r)
	m := h.measure(w, r)

	e := l.WithLevel(m.logLevel()).EmbedObject(m).Str("method", r.Method).Object("trace", t)

	if m.Failure != nil {
		e = e.EmbedObject(m.Failure)
	}

	e.Send()
}

func (h *actionHandler) trace(w http.ResponseWriter, r *http.Request) *traceParent {
	t, err := parseTraceParent(r.Header.Get(headerTraceParent))
	if err != nil {
		t = newTraceParent()

		w.Header().Set(headerTraceResponse, t.String())
	}

	return t
}

func (h *actionHandler) measure(w http.ResponseWriter, r *http.Request) *metrics {
	cm := httpsnoop.CaptureMetrics(h.next, w, r)

	var m *metrics

	if v, ok := r.Context().Value(metricsKey{}).(*metrics); ok {
		m = v
	} else {
		m = new(metrics)
	}

	m.Duration = cm.Duration
	m.Status = cm.Code
	m.Route = h.route

	if m.Route == "/" {
		m.Path = r.URL.Path
	}

	return m
}

func action(route string, next http.Handler) http.Handler {
	return &actionHandler{
		route: route,
		next:  next,
	}
}

func cors(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodOptions && len(r.Header.Get("Origin")) != 0 {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Expose-Headers", "*")
		w.Header().Set("Vary", "Origin")

		w.WriteHeader(http.StatusOK)

		return true
	}

	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	return false
}
