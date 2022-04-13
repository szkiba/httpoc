package httpoc

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/hashicorp/go-hclog"
	"github.com/mattn/go-colorable"
	"github.com/pior/runnable"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

const fieldPrefixJSON = "\t\n"

var levelMap = map[hclog.Level]zerolog.Level{
	hclog.NoLevel: zerolog.NoLevel,
	hclog.Trace:   zerolog.TraceLevel,
	hclog.Debug:   zerolog.DebugLevel,
	hclog.Info:    zerolog.InfoLevel,
	hclog.Warn:    zerolog.WarnLevel,
	hclog.Error:   zerolog.ErrorLevel,
}

type sinkAdapterToZerolog struct {
	log *zerolog.Logger
}

func (a *sinkAdapterToZerolog) Accept(name string, level hclog.Level, msg string, args ...interface{}) {
	l, ok := levelMap[level]
	if !ok {
		return
	}

	e := a.log.WithLevel(l)

	if len(name) != 0 {
		e.Str("@module", name)
	}

	last := len(args) - 1

	for i := 0; i < len(args); i++ {
		if i >= last {
			e.Interface(hclog.MissingKey, fmt.Sprint(args[i]))

			continue
		}

		if s, ok := args[i].(string); ok {
			if s == "timestamp" {
				i++

				continue
			}
		}

		if s, ok := args[i+1].(string); ok {
			if strings.HasPrefix(s, fieldPrefixJSON) {
				m := map[string]interface{}{}

				if err := json.Unmarshal([]byte(s[len(fieldPrefixJSON):]), &m); err == nil {
					e.Interface(fmt.Sprint(args[i]), m)
					i++

					continue
				}
			}
		}

		e.Interface(fmt.Sprint(args[i]), fmt.Sprint(args[i+1]))
		i++
	}

	e.Msg(msg)
}

func wrapLogger(l *zerolog.Logger) hclog.Logger {
	h := hclog.NewInterceptLogger(&hclog.LoggerOptions{Output: io.Discard})

	h.RegisterSink(&sinkAdapterToZerolog{log: l})

	return h
}

func useLogger(l hclog.Logger) {
	hclog.SetDefault(l)
	log.Default().SetFlags(0)
	log.SetOutput(hclog.Default().StandardWriter(&hclog.StandardLoggerOptions{ForceLevel: hclog.Debug}))
}

func isConsole() bool {
	if os.Getenv("FORCE_CONSOLE") != "" {
		return true
	}

	fileInfo, _ := os.Stdout.Stat()

	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func initLogging(level zerolog.Level) {
	if isConsole() {
		zlog.Logger = zlog.Output(
			zerolog.ConsoleWriter{
				Out:     colorable.NewColorableStderr(),
				NoColor: color.NoColor,
			},
		)
	}

	zerolog.SetGlobalLevel(zerolog.TraceLevel)

	zlog.Logger = zlog.Logger.Level(level)
	zerolog.DefaultContextLogger = &zlog.Logger

	useLogger(wrapLogger(&zlog.Logger))
	runnable.SetLogger(new(runnableLogger))
}

type runnableLogger struct {
}

func (l *runnableLogger) Infof(format string, args ...interface{}) {
	zlog.Logger.Debug().Msgf(format, args...)
}

func (l *runnableLogger) Debugf(format string, args ...interface{}) {
	zlog.Logger.Debug().Msgf(format, args...)
}
