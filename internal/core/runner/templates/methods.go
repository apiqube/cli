package templates

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func methodToString(value any, _ ...string) (any, error) { return fmt.Sprintf("%v", value), nil }
func methodToUpper(value any, _ ...string) (any, error) {
	return strings.ToUpper(fmt.Sprintf("%v", value)), nil
}

func methodToLower(value any, _ ...string) (any, error) {
	return strings.ToLower(fmt.Sprintf("%v", value)), nil
}

func methodTrimSpace(value any, _ ...string) (any, error) {
	return strings.TrimSpace(fmt.Sprint(value)), nil
}

func methodReplace(value any, args ...string) (any, error) {
	if len(args) < 2 {
		return value, nil
	}

	return strings.ReplaceAll(fmt.Sprintf("%v", value), clearArg(args[0]), clearArg(args[1])), nil
}

func methodPadLeft(value any, args ...string) (any, error) {
	if len(args) < 2 {
		return value, nil
	}
	n := 0
	_, _ = fmt.Sscanf(args[0], "%d", &n)
	s := fmt.Sprintf("%v", value)
	for len(s) < n {
		s = args[1] + s
	}
	return s, nil
}

func methodPadRight(value any, args ...string) (any, error) {
	if len(args) < 2 {
		return value, nil
	}
	n := 0
	_, _ = fmt.Sscanf(args[0], "%d", &n)
	s := fmt.Sprintf("%v", value)
	for len(s) < n {
		s = s + args[1]
	}
	return s, nil
}

func methodSubstring(value any, args ...string) (any, error) {
	s := fmt.Sprint(value)
	start, end := 0, len(s)
	if len(args) > 0 {
		_, _ = fmt.Sscanf(args[0], "%d", &start)
	}
	if len(args) > 1 {
		_, _ = fmt.Sscanf(args[1], "%d", &end)
	}
	if start < 0 {
		start = 0
	}
	if end > len(s) {
		end = len(s)
	}
	if start > end {
		start, end = end, start
	}
	return s[start:end], nil
}

func methodCapitalize(value any, _ ...string) (any, error) {
	s := fmt.Sprint(value)
	if len(s) == 0 {
		return s, nil
	}
	return strings.ToUpper(s[:1]) + s[1:], nil
}

func methodReverse(value any, _ ...string) (any, error) {
	s := fmt.Sprint(value)
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r), nil
}

func methodRandomCase(value any, _ ...string) (any, error) {
	s := fmt.Sprint(value)
	out := make([]rune, len(s))
	for i, c := range s {
		if i%2 == 0 {
			out[i] = []rune(strings.ToUpper(string(c)))[0]
		} else {
			out[i] = []rune(strings.ToLower(string(c)))[0]
		}
	}
	return string(out), nil
}

func methodSnakeCase(value any, _ ...string) (any, error) {
	s := fmt.Sprint(value)
	return strings.ReplaceAll(strings.ToLower(s), " ", "_"), nil
}

func methodCamelCase(value any, _ ...string) (any, error) {
	s := fmt.Sprint(value)
	parts := strings.Fields(s)
	for i := range parts {
		if i == 0 {
			parts[i] = strings.ToLower(parts[i])
		} else {
			caser := cases.Title(language.English)
			caser.String(parts[i])
		}
	}
	return strings.Join(parts, ""), nil
}

func methodSplit(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return value, nil
	}

	return strings.Split(fmt.Sprint(value), clearArg(args[0])), nil
}

func methodJoin(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return value, nil
	}

	if elems, ok := value.([]string); ok {
		return strings.Join(elems, clearArg(args[0])), nil
	}

	return value, nil
}

func methodIndex(value any, args ...string) (any, error) {
	if len(args) < 1 {
		return value, nil
	}

	if elems, ok := value.([]string); ok {
		idx, err := strconv.Atoi(args[0])
		if err != nil || idx >= len(elems) || idx < 0 {
			return value, nil
		}

		return elems[idx], nil
	}

	return value, nil
}

func methodCut(value any, args ...string) (any, error) {
	if len(args)%2 != 0 {
		return value, nil
	}

	if elems, ok := value.([]string); ok {
		startStr := clearArg(args[0])
		endStr := clearArg(args[1])

		start, err := strconv.Atoi(startStr)
		if err != nil {
			return value, err
		}

		end, err := strconv.Atoi(endStr)
		if err != nil {
			return value, err
		}

		if start < 0 || end < 0 {
			start = 0
			end = 1
		}

		if start > end {
			start, end = end, start
		}

		if len(elems) < end {
			return value, nil
		}

		return elems[start:end], nil
	}

	return value, nil
}

func methodToInt(value any, _ ...string) (any, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i, nil
		}
	}
	return value, fmt.Errorf("cannot convert %v to int", value)
}

func methodToUint(value any, _ ...string) (any, error) {
	switch v := value.(type) {
	case uint:
		return v, nil
	case int:
		if v < 0 {
			return value, fmt.Errorf("cannot convert negative int %v to uint", v)
		}
		return uint(v), nil
	case float64:
		if v < 0 {
			return value, fmt.Errorf("cannot convert negative float %v to uint", v)
		}
		return uint(v), nil
	case string:
		if u, err := strconv.ParseUint(v, 10, 64); err == nil {
			return uint(u), nil
		}
	}
	return value, fmt.Errorf("cannot convert %v to uint", value)
}

func methodToFloat(value any, _ ...string) (any, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, nil
		}
	}
	return value, fmt.Errorf("cannot convert %v to float64", value)
}

func methodToBool(value any, _ ...string) (any, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		if b, err := strconv.ParseBool(v); err == nil {
			return b, nil
		}
	case int:
		return v != 0, nil
	case float64:
		return v != 0, nil
	}
	return value, fmt.Errorf("cannot convert %v to bool", value)
}

func methodToArray(value any, _ ...string) (any, error) {
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		result := make([]any, val.Len())
		for i := 0; i < val.Len(); i++ {
			result[i] = val.Index(i).Interface()
		}
		return result, nil
	}
	return value, fmt.Errorf("cannot convert %v to array", value)
}

func clearArg(arg string) string {
	return strings.TrimLeft(strings.TrimRight(arg, "'"), "'")
}
