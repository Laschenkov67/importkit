// Package validate содержит регистр валидаторов значений после преобразования.
package validate

import (
	"fmt"
	"strings"
	"sync"
)

// Validator проверяет значение. Возвращает ошибку, если значение невалидно.
type Validator func(value any, arg string) error

var (
	mu       sync.RWMutex
	registry = map[string]Validator{}
)

func Register(name string, v Validator) error {
	mu.Lock()
	defer mu.Unlock()
	if _, dup := registry[name]; dup {
		return fmt.Errorf("validate: %q already registered", name)
	}
	registry[name] = v
	return nil
}

func Get(name string) (Validator, bool) {
	mu.RLock()
	defer mu.RUnlock()
	v, ok := registry[name]
	return v, ok
}

func Parse(spec string) (name, arg string) {
	if i := strings.IndexByte(spec, '='); i >= 0 {
		return spec[:i], spec[i+1:]
	}
	return spec, ""
}
