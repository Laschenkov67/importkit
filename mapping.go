package importkit

import (
	"fmt"

	"github.com/laschenkov67/importkit/internal/coerce"
	"github.com/laschenkov67/importkit/source"
	"github.com/laschenkov67/importkit/transform"
	"github.com/laschenkov67/importkit/validate"
)

// mapRecord применяет все маппинги к сырой записи. Возвращает Record
// или *RowError при первой ошибке (если StrictMode), иначе аккумулирует.
func (i *Importer) mapRecord(row int, raw source.RawRecord) (Record, error) {
	rec := make(Record, len(i.cfg.Mappings))
	for _, m := range i.cfg.Mappings {
		rawVal, ok := raw[m.Source]
		var value any = rawVal
		if !ok || rawVal == "" {
			switch {
			case m.Default != nil:
				value = m.Default
			case m.Required:
				return nil, &RowError{Row: row, Field: m.Source, Err: ErrFieldRequired}
			default:
				continue
			}
		}
		// Transformers.
		for _, spec := range m.Transform {
			name, arg := transform.Parse(spec)
			fn, ok := transform.Get(name)
			if !ok {
				return nil, &RowError{Row: row, Field: m.Source,
					Err: fmt.Errorf("%w: %s", ErrTransformNotFound, name)}
			}
			var err error
			value, err = fn(value, arg)
			if err != nil {
				return nil, &RowError{Row: row, Field: m.Source, Err: err}
			}
		}
		// Coerce.
		coerced, err := coerceValue(value, m.Type, m.Format)
		if err != nil {
			return nil, &RowError{Row: row, Field: m.Source, Err: err}
		}
		// Validators.
		for _, spec := range m.Validate {
			name, arg := validate.Parse(spec)
			fn, ok := validate.Get(name)
			if !ok {
				return nil, &RowError{Row: row, Field: m.Source,
					Err: fmt.Errorf("%w: %s", ErrValidatorNotFound, name)}
			}
			if err := fn(coerced, arg); err != nil {
				return nil, &RowError{Row: row, Field: m.Source, Err: err}
			}
		}
		rec[m.Target] = coerced
	}
	return rec, nil
}

func coerceValue(v any, t FieldType, format string) (any, error) {
	switch t {
	case "", TypeString:
		return coerce.String(v)
	case TypeInt:
		return coerce.Int(v)
	case TypeFloat:
		return coerce.Float(v)
	case TypeBool:
		return coerce.Bool(v)
	case TypeDate:
		return coerce.Date(v, format)
	}
	return nil, fmt.Errorf("unknown type %q", t)
}
