package importkit

import "fmt"

// Record — нормализованная запись после маппинга.
// Ключи — целевые поля (Target из FieldMapping), значения — приведённые типы.
type Record map[string]any

// String возвращает Go-репрезентацию записи (для отладки).
func (r Record) String() string { return fmt.Sprintf("%#v", map[string]any(r)) }

// Get возвращает значение поля и флаг существования.
func (r Record) Get(key string) (any, bool) { v, ok := r[key]; return v, ok }

// Result — единица потока импорта. Содержит запись или ошибку, а также
// номер исходной строки в источнике (1-based).
type Result struct {
	Row    int
	Record Record
	Err    error
}
