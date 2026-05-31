package importkit

import (
	"context"
	"errors"
	"io"

	"github.com/laschenkov67/importkit/source"
)

// Importer — основной фасад библиотеки.
// Безопасен для повторного использования: одна сущность Importer
// может выполнять Import(...) последовательно для разных Reader'ов.
// Параллельные вызовы Import должны выполняться над разными экземплярами.
type Importer struct {
	cfg *Config
}

// New создаёт Importer на основе валидного конфига.
func New(cfg *Config) (*Importer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Importer{cfg: cfg}, nil
}

// Import запускает потоковый импорт. Канал закрывается после io.EOF
// либо при cancel(ctx). Если SkipOnError=false — ошибочные записи
// тоже попадают в канал с Err != nil, но поток продолжается.
// Если StrictMode=true — при первой ошибке канал закрывается.
func (i *Importer) Import(ctx context.Context, r io.Reader) (<-chan Result, error) {
	src, err := i.newSource()
	if err != nil {
		return nil, err
	}
	if err := src.Open(ctx, r); err != nil {
		return nil, err
	}

	out := make(chan Result, 64)
	go func() {
		defer close(out)
		defer src.Close()
		var row int
		for {
			if err := ctx.Err(); err != nil {
				out <- Result{Row: row, Err: err}
				return
			}
			row++
			raw, err := src.Next(ctx)
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				out <- Result{Row: row, Err: err}
				if i.cfg.StrictMode {
					return
				}
				continue
			}
			rec, err := i.mapRecord(row, raw)
			if err != nil {
				if i.cfg.SkipOnError {
					continue
				}
				out <- Result{Row: row, Err: err}
				if i.cfg.StrictMode {
					return
				}
				continue
			}
			select {
			case out <- Result{Row: row, Record: rec}:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, nil
}

// ImportAll — удобная синхронная обёртка. Не использовать для больших файлов.
func (i *Importer) ImportAll(ctx context.Context, r io.Reader) ([]Record, []error) {
	ch, err := i.Import(ctx, r)
	if err != nil {
		return nil, []error{err}
	}
	var recs []Record
	var errs []error
	for res := range ch {
		if res.Err != nil {
			errs = append(errs, res.Err)
			continue
		}
		recs = append(recs, res.Record)
	}
	return recs, errs
}

func (i *Importer) newSource() (source.Source, error) {
	o := i.cfg.Source
	switch i.cfg.Format {
	case FormatCSV:
		d := ','
		if o.Delimiter != "" {
			d = []rune(o.Delimiter)[0]
		}
		return source.NewCSV(d, o.HasHeader), nil
	case FormatXLSX:
		return source.NewXLSX(o.Sheet, o.HeaderRow, o.DataStart), nil
	case FormatXML:
		return source.NewXML(o.ItemElement), nil
	case FormatYML:
		return source.NewYML(), nil
	case FormatOneC:
		return source.NewOneC(), nil
	}
	return nil, ErrUnknownFormat
}
