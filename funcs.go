package query

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type Func func(interface{}) (interface{}, error)

var funcnames = map[string]func(vs []interface{}) Func{
	"lshift":  leftShift,
	"rshift":  rightShift,
	"and":     and,
	"or":      or,
	"pow":     pow,
	"abs":     abs,
	"ltrim":   trimLeft,
	"rtrim":   trimRight,
	"lower":   toLower,
	"upper":   toUpper,
	"yearday": yearDay,
	"year":    year,
	"length":  length,
}

func leftShift(args []interface{}) Func {
	return func(ifi interface{}) (interface{}, error) {
		if err := checkLength(1, args); err != nil {
			return nil, fmt.Errorf("lshift: %w", err)
		}
		value, err := toInt(ifi)
		if err != nil {
			return nil, err
		}
		count, err := toInt(args[0])
		if err == nil {
			value <<= count
		}
		return value, err
	}
}

func rightShift(args []interface{}) Func {
	return func(ifi interface{}) (interface{}, error) {
		if err := checkLength(1, args); err != nil {
			return nil, fmt.Errorf("rshift: %w", err)
		}
		value, err := toInt(ifi)
		if err != nil {
			return nil, err
		}
		count, err := toInt(args[0])
		if err == nil {
			value >>= count
		}
		return value, err
	}
}

func and(args []interface{}) Func {
	return func(ifi interface{}) (interface{}, error) {
		if err := checkLength(1, args); err != nil {
			return nil, fmt.Errorf("and: %w", err)
		}
		value, err := toInt(ifi)
		if err != nil {
			return nil, err
		}
		count, err := toInt(args[0])
		if err == nil {
			value &= count
		}
		return value, err
	}
}

func or(args []interface{}) Func {
	return func(ifi interface{}) (interface{}, error) {
		if err := checkLength(1, args); err != nil {
			return nil, fmt.Errorf("or: %w", err)
		}
		value, err := toInt(ifi)
		if err != nil {
			return nil, err
		}
		count, err := toInt(args[0])
		if err == nil {
			value |= count
		}
		return value, err
	}
}

func pow(args []interface{}) Func {
	return func(ifi interface{}) (interface{}, error) {
		if err := checkLength(1, args); err != nil {
			return nil, fmt.Errorf("pow: %w", err)
		}
		exp, err := toFloat(args[0])
		if err != nil {
			return nil, err
		}
		switch value := ifi.(type) {
		case float64:
			return math.Pow(value, exp), nil
		case int64:
			val := math.Pow(float64(value), exp)
			return int64(val), nil
		default:
			return nil, castError("number", ifi)
		}
	}
}

func abs(args []interface{}) Func {
	return func(ifi interface{}) (interface{}, error) {
		if err := checkLength(0, args); err != nil {
			return nil, fmt.Errorf("abs: %w", err)
		}
		switch value := ifi.(type) {
		case float64:
			return math.Abs(value), nil
		case int64:
			val := math.Abs(float64(value))
			return int64(val), nil
		default:
			return nil, castError("number", ifi)
		}
	}
}

func yearDay(args []interface{}) Func {
	return func(ifi interface{}) (interface{}, error) {
		if err := checkLength(0, args); err != nil {
			return nil, fmt.Errorf("yearday: %w", err)
		}
		value, err := toTime(ifi)
		if err != nil {
			return nil, err
		}
		return value.YearDay(), nil
	}
}

func year(args []interface{}) Func {
	return func(ifi interface{}) (interface{}, error) {
		if err := checkLength(0, args); err != nil {
			return nil, fmt.Errorf("year: %w", err)
		}
		value, err := toTime(ifi)
		if err != nil {
			return nil, err
		}
		return value.Year(), nil
	}
}

func toLower(args []interface{}) Func {
	return func(ifi interface{}) (interface{}, error) {
		if err := checkLength(0, args); err != nil {
			return nil, fmt.Errorf("lower: %w", err)
		}
		str, err := toString(ifi)
		if err == nil {
			str = strings.ToLower(str)
		}
		return str, err
	}
}

func toUpper(args []interface{}) Func {
	return func(ifi interface{}) (interface{}, error) {
		if err := checkLength(0, args); err != nil {
			return nil, fmt.Errorf("upper: %w", err)
		}
		str, err := toString(ifi)
		if err == nil {
			str = strings.ToUpper(str)
		}
		return str, err
	}
}

func trimLeft(args []interface{}) Func {
	return func(ifi interface{}) (interface{}, error) {
		if err := checkLength(2, args); err != nil {
			return nil, fmt.Errorf("ltrim: %w", err)
		}
		str, err := toString(ifi)
		if err != nil {
			return nil, err
		}
		left, err := toString(args[0])
		if err != nil {
			return nil, err
		}
		long, err := toBool(args[1])
		if err != nil {
			return nil, err
		}
		for i := 0; strings.HasPrefix(str, left); i++ {
			if long && i > 0 {
				break
			}
			str = strings.TrimPrefix(str, left)
		}
		return str, nil
	}
}

func trimRight(args []interface{}) Func {
	return func(ifi interface{}) (interface{}, error) {
		if err := checkLength(2, args); err != nil {
			return nil, fmt.Errorf("rtrim: %w", err)
		}
		str, err := toString(ifi)
		if err != nil {
			return nil, err
		}
		right, err := toString(args[0])
		if err != nil {
			return nil, err
		}
		long, err := toBool(args[1])
		if err != nil {
			return nil, err
		}
		for i := 0; strings.HasSuffix(str, right); i++ {
			if long && i > 0 {
				break
			}
			str = strings.TrimSuffix(str, right)
		}
		return str, nil
	}
}

func length(args []interface{}) Func {
	return func(ifi interface{}) (interface{}, error) {
		if err := checkLength(0, args); err != nil {
			return nil, fmt.Errorf("length: %w", err)
		}
		switch ifi := ifi.(type) {
		case string:
			return len(ifi), nil
		case []interface{}:
			return len(ifi), nil
		case map[string]interface{}:
			return len(ifi), nil
		default:
			return nil, fmt.Errorf("length can not be applied on boolean/number")
		}
	}
}

func checkLength(want int, args []interface{}) error {
	if len(args) != want {
		return fmt.Errorf("invalid number of arguments (want %d, got %d)", want, len(args))
	}
	return nil
}

func toInt(ifi interface{}) (int64, error) {
	i, ok := ifi.(int64)
	if !ok {
		return 0, castError("int", ifi)
	}
	return i, nil
}

func toFloat(ifi interface{}) (float64, error) {
	i, ok := ifi.(float64)
	if !ok {
		return 0, castError("float", ifi)
	}
	return i, nil
}

func toBool(ifi interface{}) (bool, error) {
	i, ok := ifi.(bool)
	if !ok {
		return i, castError("bool", ifi)
	}
	return i, nil
}

func toString(ifi interface{}) (string, error) {
	i, ok := ifi.(string)
	if !ok {
		return i, castError("string", ifi)
	}
	return i, nil
}

func toTime(ifi interface{}) (time.Time, error) {
	i, ok := ifi.(time.Time)
	if !ok {
		return i, castError("time", ifi)
	}
	return i, nil
}
