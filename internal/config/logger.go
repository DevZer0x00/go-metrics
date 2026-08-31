package config

import (
	"io"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitLog(w io.Writer) {
	log.Logger = zerolog.New(w).
		With().
		Timestamp().
		Caller().
		Logger().
		Level(zerolog.InfoLevel)
}
