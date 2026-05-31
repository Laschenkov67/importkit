// Package transform содержит цепочки преобразований сырых значений
// перед валидацией. Все встроенные трансформеры безопасны для nil-значений.
package transform

import (
	"fmt"
	"strings"
	"sync"
)

// Transformer преобразует значение. arg — параметр после '=' в декларации.
type Transformer func(value any, arg string) (any, error)

var (
	mu       sync.RWMutex
	registry = map[string]Transformer{}
)

// Register добавляет именованный трансформер. Перерегистрация запрещена.
func Register(name string, t Transformer) error {
	mu.Lock()
	defer mu.Unlock()
	if _, dup := registry[name]; dup {
		return fmt.Errorf("transform: %q already registered", name)
	}
	registry[name] = t
	return nil
}

// Get возвращает трансформер по имени.
func Get(name string) (Transformer, bool) {
	mu.RLock()
	defer mu.RUnlock()
	t, ok := registry[name]
	return t, ok
}

// Parse разбирает строку вида "name" или "name=arg".
func Parse(spec string) (name, arg string) {
	if i := strings.IndexByte(spec, '='); i >= 0 {
		return spec[:i], spec[i+1:]
	}
	return spec, ""
}
