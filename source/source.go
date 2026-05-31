package source

import (
	"context"
	"io"
)

// RawRecord — «сырая» запись из источника: ключ = имя колонки/тега, значение = строка.
type RawRecord map[string]string

// Source — абстракция над любым форматом-источником.
// Реализация ДОЛЖНА быть потокобезопасной только в рамках одного вызова Read.
type Source interface {
	// Open подготавливает источник к чтению.
	Open(ctx context.Context, r io.Reader) error
	// Next возвращает следующую сырую запись. io.EOF — корректный конец.
	Next(ctx context.Context) (RawRecord, error)
	// Close освобождает ресурсы.
	Close() error
}

// Factory создаёт Source по опциям.
type Factory func(opts map[string]any) (Source, error)
