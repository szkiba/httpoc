package httpoc

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/pior/runnable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Engine struct {
	http.ServeMux
	*http.Server
	dependency []runnable.Runnable
	dependent  []runnable.Runnable
}

var DefaultEngine = new(Engine)

func Handle(pattern string, handler http.Handler) {
	DefaultEngine.Handle(pattern, handler)
}

func HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	DefaultEngine.Handle(pattern, http.HandlerFunc(handler))
}

func Dependency(runners ...runnable.Runnable) {
	DefaultEngine.Dependency(runners...)
}

func Dependent(runners ...runnable.Runnable) {
	DefaultEngine.Dependent(runners...)
}

func Run(conf interface{}) {
	DefaultEngine.Run(conf)
}

func (e *Engine) Handle(pattern string, handler http.Handler) {
	e.ServeMux.Handle(pattern, action(pattern, handler))
}

func (e *Engine) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	e.Handle(pattern, http.HandlerFunc(handler))
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if cors(w, r) {
		return
	}

	defer func() {
		if rec := recover(); rec != nil {
			stack := make([]byte, 8096)
			stack = stack[:runtime.Stack(stack, false)]

			log.Ctx(r.Context()).Error().Bytes("stack", stack).Str("error", "panic").Interface("error_description", rec).Send()

			WriteFailure(w, r, ErrServerError)
		}
	}()

	e.ServeMux.ServeHTTP(w, r)
}

func (e *Engine) Dependency(runners ...runnable.Runnable) {
	e.dependency = append(e.dependency, runners...)
}

func (e *Engine) Dependent(runners ...runnable.Runnable) {
	e.dependent = append(e.dependent, runners...)
}

func (e *Engine) Run(conf interface{}) {
	opts := configure(conf)

	e.Server = &http.Server{
		Handler:      &e.ServeMux,
		Addr:         fmt.Sprintf(":%d", opts.Port),
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
		IdleTimeout:  idleTimeout,
	}

	g := runnable.Manager(&runnable.ManagerOptions{ShutdownTimeout: gracePeriod})

	srv := runnable.HTTPServer(e.Server)
	g.Add(srv, e.dependency...)

	for _, d := range e.dependent {
		g.Add(d, srv)
	}

	if c, ok := conf.(zerolog.LogObjectMarshaler); ok {
		log.Info().EmbedObject(c).Msg("Starting server")
	}

	runnable.Run(g.Build())
}

const (
	writeTimeout = time.Second * 2
	readTimeout  = time.Second * 2
	idleTimeout  = time.Second * 20
	gracePeriod  = 5 * time.Second
)
