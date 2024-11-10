package filestorage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewStorage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s, _ := NewStorage(ctx, "test", false, 1)

	assert.Equal(t, "test", s.FilePath)
	// чтобы воркеры внутри успели запуститься
	time.Sleep(time.Millisecond * 100)
}
