package sink

import (
	"context"

	"github.com/laschenkov67/importkit"
)

// SliceSink аккумулирует все успешные Record в памяти.
type SliceSink struct {
	Records []importkit.Record
	Errors  []error
}

func (s *SliceSink) Consume(ctx context.Context, ch <-chan importkit.Result) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case r, ok := <-ch:
			if !ok {
				return nil
			}
			if r.Err != nil {
				s.Errors = append(s.Errors, r.Err)
				continue
			}
			s.Records = append(s.Records, r.Record)
		}
	}
}
