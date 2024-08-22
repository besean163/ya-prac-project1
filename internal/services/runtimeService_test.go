package services

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRetryFunc(t *testing.T) {

	testCases := []struct {
		Name             string
		TryCount         int
		MustMakeAttempts int
		Error            error
	}{
		{
			Name:             "0 attempts",
			TryCount:         0,
			MustMakeAttempts: 0,
			Error:            nil,
		},
		{
			Name:             "3 attempts without error",
			TryCount:         3,
			MustMakeAttempts: 1,
			Error:            nil,
		},
		{
			Name:             "3 attempts with error",
			TryCount:         3,
			MustMakeAttempts: 3,
			Error:            errors.New("connection refused"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			retry := getRetryFunc(tc.TryCount, 0, 0)
			count := 0
			for retry(tc.Error) {
				count++
			}

			assert.Equal(t, tc.MustMakeAttempts, count)

			assert.False(t, retry(tc.Error))

		})

	}
}
