package validate

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

func init() {
	_ = Register("required", required)
	_ = Register("min", minFn)
	_ = Register("max", maxFn)
	_ = Register("len_min", lenMinFn)
	_ = Register("len_max", lenMaxFn)
	_ = Register("regex", regexFn)
	_ = Register("in", inFn) // arg = "a,b,c"
}

func required(v any, _ string) error {
	if v == nil {
		return errors.New("required")
	}
	if s, ok := v.(string); ok && s == "" {
		return errors.New("required")
	}
	return nil
}

func toFloat(v any) (float64, error) {
	switch x := v.(type) {
	case float64:
		return x, nil
	case float32:
		return float64(x), nil
	case int:
		return float64(x), nil
	case int64:
		return float64(x), nil
	case string:
		return strconv.ParseFloat(strings.ReplaceAll(x, ",", "."), 64)
	}
	return 0, fmt.Errorf("not a number: %T", v)
}

func minFn(v any, arg string) error {
	bound, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		return err
	}
	f, err := toFloat(v)
	if err != nil {
		return err
	}
	if f < bound {
		return fmt.Errorf("value %v < min %v", f, bound)
	}
	return nil
}

func maxFn(v any, arg string) error {
	bound, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		return err
	}
	f, err := toFloat(v)
	if err != nil {
		return err
	}
	if f > bound {
		return fmt.Errorf("value %v > max %v", f, bound)
	}
	return nil
}

func lenMinFn(v any, arg string) error {
	n, err := strconv.Atoi(arg)
	if err != nil {
		return err
	}
	s, ok := v.(string)
	if !ok {
		return errors.New("len_min: not a string")
	}
	if len([]rune(s)) < n {
		return fmt.Errorf("len %d < %d", len([]rune(s)), n)
	}
	return nil
}

func lenMaxFn(v any, arg string) error {
	n, err := strconv.Atoi(arg)
	if err != nil {
		return err
	}
	s, ok := v.(string)
	if !ok {
		return errors.New("len_max: not a string")
	}
	if len([]rune(s)) > n {
		return fmt.Errorf("len %d > %d", len([]rune(s)), n)
	}
	return nil
}

var (
	reCache   = map[string]*regexp.Regexp{}
	reCacheMu sync.RWMutex
)

func regexFn(v any, arg string) error {
	reCacheMu.RLock()
	re, ok := reCache[arg]
	reCacheMu.RUnlock()
	if !ok {
		var err error
		re, err = regexp.Compile(arg)
		if err != nil {
			return err
		}
		reCacheMu.Lock()
		reCache[arg] = re
		reCacheMu.Unlock()
	}
	s, ok2 := v.(string)
	if !ok2 {
		return errors.New("regex: not a string")
	}
	if !re.MatchString(s) {
		return fmt.Errorf("regex %q not matched", arg)
	}
	return nil
}

func inFn(v any, arg string) error {
	s := fmt.Sprint(v)
	for _, x := range strings.Split(arg, ",") {
		if strings.TrimSpace(x) == s {
			return nil
		}
	}
	return fmt.Errorf("value %q not in [%s]", s, arg)
}
