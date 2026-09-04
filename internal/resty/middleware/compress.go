package middleware

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"go-metrics/internal/model"

	"github.com/rs/zerolog/log"
	"resty.dev/v3"
)

func RequestGzipCompress(_ *resty.Client, request *resty.Request) error {
	var bodyBytes []byte
	var err error

	switch b := request.Body.(type) {
	case []byte:
		bodyBytes = b
	case string:
		bodyBytes = []byte(b)
	case *model.Metric:
		bodyBytes, err = json.Marshal(b)
		if err != nil {
			log.Error().Err(err).Msg("failed to json marshal metric")
			return nil
		}
	default:
		return nil
	}

	var buff bytes.Buffer

	gzipWriter := gzip.NewWriter(&buff)
	_, err = gzipWriter.Write(bodyBytes)
	if err != nil {
		log.Error().Err(err).Any("body", bodyBytes).Msg("failed to compress request body")
		return nil
	}

	err = gzipWriter.Close()
	if err != nil {
		log.Error().Err(err).Any("body", bodyBytes).Msg("failed to compress request body")
		return nil
	}

	request.SetHeader("Content-Encoding", "gzip")
	request.SetBody(buff.Bytes())

	return nil
}
