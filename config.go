package httpoc

import (
	"flag"
	"os"

	"github.com/jinzhu/copier"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	loadDotEnv(".env.local")
	loadDotEnv(".env")
}

func loadDotEnv(filename string) {
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		if err := godotenv.Load(filename); err != nil {
			log.Fatal().Err(err).Send()
		}
	}
}

type options struct {
	Port     int
	LogLevel zerolog.Level
	App      string
}

func configure(conf interface{}) *options {
	initLogging(zerolog.ErrorLevel)

	opts := new(options)

	if err := copier.Copy(opts, conf); err != nil {
		log.Fatal().Err(err).Send()
	}

	help := flag.Bool("help", false, "print usage")

	flag.Parse()

	if *help {
		err := envconfig.Usagef(opts.App, conf, os.Stderr, usageFormat)
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		os.Exit(0)
	}

	if err := envconfig.Process(opts.App, conf); err != nil {
		log.Fatal().Err(err).Send()
	}

	if err := copier.Copy(opts, conf); err != nil {
		log.Fatal().Err(err).Send()
	}

	initLogging(opts.LogLevel)

	return opts
}

const usageFormat = `This application is configured via the environment.

The following environment variables can be used:
{{range .}}
{{usage_key .}}
{{- if usage_default .}}
  default: {{usage_default .}}
{{end -}}
{{- if usage_description .}}
  {{usage_description .}}
{{end -}}
{{end}}
`
