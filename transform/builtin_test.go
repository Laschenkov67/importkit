package transform_test

import (
	"testing"

	"github.com/laschenkov67/importkit/transform"
)

func TestBuiltinTransforms(t *testing.T) {
	cases := []struct {
		name, spec string
		in, want   any
	}{
		{"trim", "trim", "  hi  ", "hi"},
		{"lower", "lower", "ABC", "abc"},
		{"upper", "upper", "abc", "ABC"},
		{"default empty", "default=N/A", "", "N/A"},
		{"default keep", "default=N/A", "x", "x"},
		{"replace", "replace= |_", "Hello World", "Hello_World"},
		{"only_digits", "only_digits", "a1b2c3", "123"},
		{"prefix", "prefix=SKU-", "001", "SKU-001"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			name, arg := transform.Parse(c.spec)
			fn, ok := transform.Get(name)
			if !ok {
				t.Fatalf("no transformer %q", name)
			}
			got, err := fn(c.in, arg)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}
