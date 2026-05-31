// Package sink содержит готовые приёмники потока импорта.
package sink

import (
	"context"

	"github.com/laschenkov67/importkit"
)

// Sink принимает поток Result и обрабатывает его как угодно.
type Sink interface {
	Consume(ctx context.Context, ch <-chan importkit.Result) error
}
