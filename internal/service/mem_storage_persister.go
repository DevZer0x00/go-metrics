package service

import (
	"encoding/json"
	"fmt"
	"go-metrics/internal/model"
	"go-metrics/internal/repository"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

type MemMetricsPersister struct {
	storage         *repository.MemStorage
	loadOnInit      bool
	storageFilePath string
	logger          *zerolog.Logger
}

func (m *MemMetricsPersister) Init() error {
	storageFile, err := os.OpenFile(m.storageFilePath, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		return fmt.Errorf("open storage file: %w", err)
	}

	defer func(storageFile *os.File) {
		err = storageFile.Close()
		if err != nil {
			m.logger.Error().Err(err).Msgf("close storage file")
		}
	}(storageFile)

	if !m.loadOnInit {
		return nil
	}

	storageData, err := io.ReadAll(storageFile)
	if err != nil {
		return fmt.Errorf("read storage file: %w", err)
	}

	metrics := make([]model.Metric, 0)
	if len(storageData) != 0 {
		err = json.Unmarshal(storageData, &metrics)
		if err != nil {
			m.logger.Info().Str("content", string(storageData)).Msgf("storage file content")
			return fmt.Errorf("unmarshal storage file: %w", err)
		}
	}

	for _, metric := range metrics {
		err = m.storage.Save(&metric)
		if err != nil {
			return fmt.Errorf("save metric: %w", err)
		}
	}

	return nil
}

func (m *MemMetricsPersister) Flush() error {
	tmpFilePath := m.storageFilePath + ".tmp"
	storageFile, err := os.OpenFile(tmpFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("open storage file: %w", err)
	}

	defer func(storageFile *os.File) {
		err = storageFile.Close()
		if err != nil {
			m.logger.Error().Err(err).Msgf("close storage file")
		}
	}(storageFile)

	metrics, err := m.storage.All()
	if err != nil {
		return fmt.Errorf("get all metrics: %w", err)
	}

	lines := make([]string, len(metrics))
	for i, metric := range metrics {
		encoded, err := json.Marshal(*metric)
		if err != nil {
			return fmt.Errorf("marshal metric: %w", err)
		}

		lines[i] = "  " + string(encoded)
	}

	dbData := fmt.Sprintf("[\n%s\n]", strings.Join(lines, ",\n"))
	_, err = storageFile.Write([]byte(dbData))
	if err != nil {
		return fmt.Errorf("write metrics: %w", err)
	}
	err = storageFile.Sync()
	if err != nil {
		return fmt.Errorf("sync metrics: %w", err)
	}

	err = os.Rename(tmpFilePath, m.storageFilePath)
	if err != nil {
		return err
	}

	return nil
}

func NewMetricsPersister(storage *repository.MemStorage, logger *zerolog.Logger, loadOnInit bool, storageFilePath string) *MemMetricsPersister {
	return &MemMetricsPersister{
		storage:         storage,
		loadOnInit:      loadOnInit,
		storageFilePath: storageFilePath,
		logger:          logger,
	}
}
