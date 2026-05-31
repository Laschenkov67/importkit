package importkit

import (
	"github.com/laschenkov67/importkit/transform"
	"github.com/laschenkov67/importkit/validate"
)

// RegisterTransformer регистрирует пользовательский трансформер в глобальном реестре.
// Безопасно вызывать в init() пакета-расширения.
func RegisterTransformer(name string, t transform.Transformer) error {
	return transform.Register(name, t)
}

// RegisterValidator регистрирует пользовательский валидатор.
func RegisterValidator(name string, v validate.Validator) error {
	return validate.Register(name, v)
}
