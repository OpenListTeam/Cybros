package utils

import (
	"context"
	"sync"

	"github.com/gotd/td/tgerr"
)

var telegramAPIMu sync.Mutex

func RetryFloodWait[T any](ctx context.Context, call func() (T, error)) (T, error) {
	telegramAPIMu.Lock()
	defer telegramAPIMu.Unlock()

	for {
		value, err := call()
		if err == nil {
			return value, nil
		}

		retry, waitErr := tgerr.FloodWait(ctx, err)
		if retry {
			continue
		}

		var zero T
		return zero, waitErr
	}
}
