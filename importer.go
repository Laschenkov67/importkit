package importkit

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/laschenkov67/importkit/source"
	"github.com/laschenkov67/importkit/transform"
	"github.com/laschenkov67/importkit/validate"
)

// Importer — основной фасад библиотеки.
// Importer иммутабелен после New() и безопасен для конкурентного использования:
// один и тот же экземпляр можно использовать для нескольких одновременных
// вызовов Import(...) над разными Reader'ами — каждый вызов работает
// с собственным Source и не разделяет изменяемое состояние с другими.
type Importer struct {
	cfg *Config
}

// New создаёт Importer на основе валидного конфига. Помимо структурной
// проверки (cfg.Validate), также проверяет, что все имена трансформеров
// и валидаторов, указанные в маппингах, зарегистрированы — это позволяет
// поймать опечатку в конфиге сразу, а не на первой строке данных в проде,
// тем более что при SkipOnError=true такая ошибка иначе осталась бы незамеченной.
func New(cfg *Config) (*Importer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if err := validateRegistries(cfg); err != nil {
		return nil, err
	}
	return &Importer{cfg: cfg}, nil
}

func validateRegistries(cfg *Config) error {
	for _, m := range cfg.Mappings {
		for _, spec := range m.Transform {
			name, _ := transform.Parse(spec)
			if _, ok := transform.Get(name); !ok {
				return fmt.Errorf("%w: field %q: transformer %q not registered",
					ErrConfigInvalid, m.Target, name)
			}
		}
		for _, spec := range m.Validate {
			name, _ := validate.Parse(spec)
			if _, ok := validate.Get(name); !ok {
				return fmt.Errorf("%w: field %q: validator %q not registered",
					ErrConfigInvalid, m.Target, name)
			}
		}
	}
	return nil
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
		_ = src.Close() // освобождаем ресурсы, которые Source мог успеть захватить до ошибки
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
	var recs []Record //nolint:prealloc // итоговый размер заранее неизвестен — он определяется потоком из канала
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
