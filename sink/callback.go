package sink

import (
	"context"

	"github.com/laschenkov67/importkit"
)

// CallbackSink вызывает функцию на каждую запись. Если cb возвращает ошибку,
// Consume завершает поток.
type CallbackSink struct{ Fn func(importkit.Result) error }

func (s *CallbackSink) Consume(ctx context.Context, ch <-chan importkit.Result) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case r, ok := <-ch:
			if !ok {
				return nil
			}
			if err := s.Fn(r); err != nil {
				return err
			}
		}
	}
}
