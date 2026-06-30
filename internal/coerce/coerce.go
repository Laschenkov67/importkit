// Package coerce приводит строковые значения источников к целевым Go-типам.
package coerce

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func String(v any) (string, error) {
	if v == nil {
		return "", nil
	}
	return fmt.Sprint(v), nil
}

func Int(v any) (int64, error) {
	switch x := v.(type) {
	case int64:
		return x, nil
	case int:
		return int64(x), nil //nolint:gosec // int always fits in int64, widening conversion is safe
	case float64:
		return int64(x), nil
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return 0, nil
		}
		return strconv.ParseInt(s, 10, 64)
	}
	return 0, fmt.Errorf("coerce: cannot convert %T to int", v)
}

func Float(v any) (float64, error) {
	switch x := v.(type) {
	case float64:
		return x, nil
	case int:
		return float64(x), nil
	case int64:
		return float64(x), nil
	case string:
		s := strings.ReplaceAll(strings.TrimSpace(x), ",", ".")
		if s == "" {
			return 0, nil
		}
		return strconv.ParseFloat(s, 64)
	}
	return 0, fmt.Errorf("coerce: cannot convert %T to float", v)
}

func Bool(v any) (bool, error) {
	switch x := v.(type) {
	case bool:
		return x, nil
	case string:
		switch strings.ToLower(strings.TrimSpace(x)) {
		case "1", "true", "y", "yes", "да", "истина":
			return true, nil
		case "", "0", "false", "n", "no", "нет", "ложь":
			return false, nil
		}
	}
	return false, fmt.Errorf("coerce: cannot convert %v to bool", v)
}

// Date парсит дату. Если layout задан явно — используется только он
// (чтобы не подменить формат и не получить неверную дату из-за случайного
// совпадения с другим layout'ом). Если layout пуст — пробуем по очереди
// RFC3339, 2006-01-02, 02.01.2006, 02/01/2006.
func Date(v any, layout string) (time.Time, error) {
	s, ok := v.(string)
	if !ok {
		return time.Time{}, fmt.Errorf("coerce: date expects string, got %T", v)
	}
	s = strings.TrimSpace(s)
	if layout != "" {
		t, err := time.Parse(layout, s)
		if err != nil {
			return time.Time{}, fmt.Errorf("coerce: cannot parse date %q with layout %q: %w", s, layout, err)
		}
		return t, nil
	}
	layouts := []string{time.RFC3339, "2006-01-02", "02.01.2006", "02/01/2006"}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("coerce: cannot parse date %q", s)
}
