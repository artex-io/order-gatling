package cmd

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"sylr.dev/fix/config"
)

func InitLogger() error {
	options := config.GetOptions()
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "Jan 2 15:04:05.000-0700",
	}
	multi := zerolog.MultiLevelWriter(consoleWriter)
	logger := zerolog.New(multi).With().Timestamp().Logger().Level(config.IntToZerologLevel(options.Verbose))

	if options.LogCaller {
		logger = logger.With().Caller().Logger()
	}
	config.SetLogger(&logger)
	return nil
}
