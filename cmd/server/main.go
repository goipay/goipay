package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/chekist32/goipay/internal/app"
	"github.com/chekist32/goipay/internal/util"
	"github.com/rs/zerolog"
)

type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
	PanicLevel LogLevel = "panic"
)

func main() {
	configPath := flag.String("config", "config.yml", "Path to the config file")
	clientCAs := flag.String("client-ca", "", "Comma-separated list of paths to client certificate authority files (for mTLS)")
	reflection := flag.Bool("reflection", false, "Enables gRPC server reflection")
	logLevel := flag.String("log-level", "", "Defines the logging level")
	flag.Parse()

	switch LogLevel(util.GetOptionOrEnvValue("LOG_LEVEL", *logLevel)) {
	case DebugLevel:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case InfoLevel:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case WarnLevel:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case ErrorLevel:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case FatalLevel:
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case PanicLevel:
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	}

	app := app.NewApp(app.CliOpts{
		ConfigPath:        *configPath,
		ClientCAPaths:     *clientCAs,
		ReflectionEnabled: *reflection,
	})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := app.Start(ctx); err != nil {
		log.Fatal(err)
	}
}
