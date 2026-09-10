package config

import (
	"io"

	"github.com/rs/zerolog"
)

func InitLog(w io.Writer) *zerolog.Logger {
	logger := zerolog.New(w).
		With().
		Timestamp().
		Caller().
		Logger().
		Level(zerolog.InfoLevel)

	return &logger
}
