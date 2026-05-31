package transform

import (
	"fmt"
	"strings"
	"unicode"
)

func init() {
	must(Register("trim", trimFn))
	must(Register("lower", lowerFn))
	must(Register("upper", upperFn))
	must(Register("replace", replaceFn)) // arg = "from|to"
	must(Register("default", defaultFn)) // arg = "значение"
	must(Register("strip_spaces", stripSpacesFn))
	must(Register("only_digits", onlyDigitsFn))
	must(Register("prefix", prefixFn)) // arg = "пре"
	must(Register("suffix", suffixFn))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func asString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

func trimFn(v any, _ string) (any, error)  { return strings.TrimSpace(asString(v)), nil }
func lowerFn(v any, _ string) (any, error) { return strings.ToLower(asString(v)), nil }
func upperFn(v any, _ string) (any, error) { return strings.ToUpper(asString(v)), nil }

func replaceFn(v any, arg string) (any, error) {
	parts := strings.SplitN(arg, "|", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("transform.replace: expected 'from|to', got %q", arg)
	}
	return strings.ReplaceAll(asString(v), parts[0], parts[1]), nil
}

func defaultFn(v any, arg string) (any, error) {
	s := asString(v)
	if s == "" {
		return arg, nil
	}
	return v, nil
}

func stripSpacesFn(v any, _ string) (any, error) {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, asString(v)), nil
}

func onlyDigitsFn(v any, _ string) (any, error) {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, asString(v)), nil
}

func prefixFn(v any, arg string) (any, error) { return arg + asString(v), nil }
func suffixFn(v any, arg string) (any, error) { return asString(v) + arg, nil }
