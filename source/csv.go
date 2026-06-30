package source

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
)

// CSVSource — реализация Source для RFC 4180 CSV с произвольным разделителем.
// Source не владеет переданным io.Reader и не закрывает его — это остаётся
// ответственностью вызывающего кода.
type CSVSource struct {
	Delimiter rune
	HasHeader bool

	reader  *csv.Reader
	headers []string
}

// NewCSV создаёт CSV-источник. delim — например ',' или ';'.
func NewCSV(delim rune, hasHeader bool) *CSVSource {
	if delim == 0 {
		delim = ','
	}
	return &CSVSource{Delimiter: delim, HasHeader: hasHeader}
}

func (s *CSVSource) Open(_ context.Context, r io.Reader) error {
	s.reader = csv.NewReader(r)
	s.reader.Comma = s.Delimiter
	s.reader.LazyQuotes = true
	s.reader.FieldsPerRecord = -1
	if s.HasHeader {
		h, err := s.reader.Read()
		if err != nil {
			return fmt.Errorf("csv: read header: %w", err)
		}
		s.headers = h
	}
	return nil
}

func (s *CSVSource) Next(ctx context.Context) (RawRecord, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	row, err := s.reader.Read()
	if err != nil {
		return nil, err
	}
	out := make(RawRecord, len(row))
	for i, v := range row {
		key := fmt.Sprintf("col_%d", i)
		if i < len(s.headers) {
			key = s.headers[i]
		}
		out[key] = v
	}
	return out, nil
}

func (s *CSVSource) Close() error {
	return nil
}
