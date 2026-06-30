package source

import (
	"context"
	"fmt"
	"io"

	"github.com/xuri/excelize/v2"
)

// XLSXSource читает Excel через excelize.
type XLSXSource struct {
	Sheet     string
	HeaderRow int // 1-based; 0 = без заголовка
	DataStart int // 1-based; 0 = сразу после заголовка

	file    *excelize.File
	rows    *excelize.Rows
	headers []string
	cursor  int
}

func NewXLSX(sheet string, headerRow, dataStart int) *XLSXSource {
	return &XLSXSource{Sheet: sheet, HeaderRow: headerRow, DataStart: dataStart}
}

func (s *XLSXSource) Open(_ context.Context, r io.Reader) error {
	// excelize требует io.Reader, но эффективнее с временным файлом, если r не *os.File.
	f, err := excelize.OpenReader(r)
	if err != nil {
		return fmt.Errorf("xlsx: open: %w", err)
	}
	s.file = f
	if s.Sheet == "" {
		s.Sheet = f.GetSheetName(0)
	}

	rows, err := f.Rows(s.Sheet)
	if err != nil {
		return fmt.Errorf("xlsx: rows: %w", err)
	}
	s.rows = rows
	s.cursor = 0

	// Прочитать заголовки.
	if s.HeaderRow > 0 {
		for s.rows.Next() {
			s.cursor++
			cols, err := s.rows.Columns()
			if err != nil {
				return err
			}
			if s.cursor == s.HeaderRow {
				s.headers = cols
				break
			}
		}
	}
	// Промотать до DataStart.
	start := s.DataStart
	if start == 0 {
		start = s.HeaderRow + 1
	}
	for s.cursor+1 < start && s.rows.Next() {
		s.cursor++
	}
	return nil
}

func (s *XLSXSource) Next(ctx context.Context) (RawRecord, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !s.rows.Next() {
		if err := s.rows.Error(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}
	s.cursor++
	cols, err := s.rows.Columns()
	if err != nil {
		return nil, err
	}
	out := make(RawRecord, len(cols))
	for i, v := range cols {
		key := fmt.Sprintf("col_%d", i)
		if i < len(s.headers) && s.headers[i] != "" {
			key = s.headers[i]
		}
		out[key] = v
	}
	return out, nil
}

func (s *XLSXSource) Close() error {
	if s.rows != nil {
		_ = s.rows.Close()
	}
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}

// Hint: для очень больших XLSX (>500MB) предпочтительно подавать *os.File,
// чтобы excelize мог использовать streaming I/O.
