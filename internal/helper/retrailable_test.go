package helper

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

type sleeper struct {
	sleepFunc func(time.Duration)
}

func (sl *sleeper) Sleep(duration time.Duration) <-chan time.Time {
	sl.sleepFunc(duration)

	return time.After(time.Duration(0))
}

func TestRetryableHelperRunError(t *testing.T) {
	counter := 0

	sl := &sleeper{
		sleepFunc: func(duration time.Duration) {
			counter++

			switch counter {
			case 1:
				assert.Equal(t, time.Duration(0), duration)
			case 2:
				assert.Equal(t, time.Second, duration)
			case 3:
				assert.Equal(t, time.Second*3, duration)
			case 4:
				assert.Equal(t, time.Second*5, duration)
			}
		},
	}

	someErr := errors.New("some error")
	logger := zerolog.New(io.Discard)

	helper := NewRetryableHelper(
		sl,
		func(err error) bool {
			return true
		},
		[]time.Duration{
			time.Second,
			time.Second * 3,
			time.Second * 5,
		},
		&logger,
	)

	err := helper.Run(t.Context(), func() error {
		return someErr
	})
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, someErr))
	assert.Equal(t, 4, counter)
}

func TestRetryableHelperRunWithoutRetry(t *testing.T) {
	counter := 0

	sl := &sleeper{
		sleepFunc: func(duration time.Duration) {
			counter++
		},
	}

	logger := zerolog.New(io.Discard)

	helper := NewRetryableHelper(
		sl,
		func(err error) bool {
			return false
		},
		[]time.Duration{
			time.Second,
			time.Second * 3,
			time.Second * 5,
		},
		&logger,
	)

	err := helper.Run(t.Context(), func() error {
		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, counter)

	counter = 0

	err = helper.Run(t.Context(), func() error {
		return errors.New("some error")
	})
	assert.NotNil(t, err)
	assert.Equal(t, 1, counter)
}
