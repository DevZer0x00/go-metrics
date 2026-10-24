package helper

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

type Sleeper interface {
	Sleep(time.Duration) <-chan time.Time
}

type StandardSleeper struct{}

func (s *StandardSleeper) Sleep(duration time.Duration) <-chan time.Time {
	return time.After(duration)
}

type RetryableHelper struct {
	sleeper             Sleeper
	triesPeriods        []time.Duration
	applicableErrorFunc func(error) bool
	logger              *zerolog.Logger
}

func (helper *RetryableHelper) Run(ctx context.Context, retryableFunc func() error) error {
	tries := append([]time.Duration{0}, helper.triesPeriods...)
	try := 0

	var err error

	for _, sleepDur := range tries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-helper.sleeper.Sleep(sleepDur):
		}

		if try > 0 {
			helper.logger.Info().Err(err).Int("try", try).Msg("retry")
		}
		err = retryableFunc()
		if err == nil {
			return nil
		}

		if helper.applicableErrorFunc(err) {
			try++
		} else {
			return err
		}
	}

	return err
}

func NewRetryableHelper(
	sleeper Sleeper,
	applicableError func(err error) bool,
	triesPeriods []time.Duration,
	logger *zerolog.Logger,
) *RetryableHelper {
	if triesPeriods == nil {
		triesPeriods = []time.Duration{
			time.Second,
			time.Second * 3,
			time.Second * 5,
		}
	}

	return &RetryableHelper{
		sleeper:             sleeper,
		triesPeriods:        triesPeriods,
		applicableErrorFunc: applicableError,
		logger:              logger,
	}
}
