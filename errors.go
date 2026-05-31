package importkit

import (
	"errors"
	"fmt"
)

var (
	ErrUnknownFormat     = errors.New("importkit: unknown source format")
	ErrConfigInvalid     = errors.New("importkit: invalid config")
	ErrSourceClosed      = errors.New("importkit: source closed")
	ErrFieldRequired     = errors.New("importkit: required field missing")
	ErrTransformNotFound = errors.New("importkit: transformer not registered")
	ErrValidatorNotFound = errors.New("importkit: validator not registered")
)

// RowError оборачивает ошибку с привязкой к строке/полю источника.
type RowError struct {
	Row   int
	Field string
	Err   error
}

func (e *RowError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("row %d, field %q: %v", e.Row, e.Field, e.Err)
	}
	return fmt.Sprintf("row %d: %v", e.Row, e.Err)
}

func (e *RowError) Unwrap() error { return e.Err }
