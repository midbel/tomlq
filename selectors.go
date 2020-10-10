package query

import (
	"fmt"
)

type Truthy struct{}

func (_ Truthy) Select(ifi interface{}) interface{} {
	if isTrue(ifi) {
		return ifi
	}
	return nil
}

type Falsy struct{}

func (_ Falsy) Select(ifi interface{}) interface{} {
	if !isTrue(ifi) {
		return ifi
	}
	return nil
}

type Int struct{}

func (_ Int) Select(ifi interface{}) interface{} {
	_, ok := ifi.(int64)
	if !ok {
		return nil
	}
	return ifi
}

type Float struct{}

func (_ Float) Select(ifi interface{}) interface{} {
	_, ok := ifi.(float64)
	if !ok {
		return nil
	}
	return ifi
}

type Number struct{}

func (_ Number) Select(ifi interface{}) interface{} {
	switch ifi.(type) {
	case int64:
	case float64:
	default:
		return nil
	}
	return ifi
}

type Boolean struct{}

func (_ Boolean) Select(ifi interface{}) interface{} {
	_, ok := ifi.(bool)
	if !ok {
		return nil
	}
	return ifi
}

type String struct{}

func (_ String) Select(ifi interface{}) interface{} {
	_, ok := ifi.(string)
	if !ok {
		return nil
	}
	return ifi
}

type First struct{}

func (_ First) Select(ifi interface{}) interface{} {
	arr, ok := ifi.([]interface{})
	if !ok || len(arr) == 0 {
		return arr
	}
	return arr[:1]
}

func (_ First) String() string {
	return ":first"
}

type Last struct{}

func (_ Last) Select(ifi interface{}) interface{} {
	arr, ok := ifi.([]interface{})
	if !ok || len(arr) == 0 {
		return arr
	}
	return arr[len(arr)-1:]
}

func (_ Last) String() string {
	return ":last"
}

type At struct {
	index int
}

func (a At) Select(ifi interface{}) interface{} {
	arr, ok := ifi.([]interface{})
	if !ok || len(arr) == 0 || a.index >= len(arr) {
		return nil
	}
	return arr[a.index : a.index+1]
}

func (a At) String() string {
	return fmt.Sprintf(":at(index: %d)", a.index)
}

type Range struct {
	start int
	end   int
}

func (r Range) Select(ifi interface{}) interface{} {
	arr, ok := ifi.([]interface{})
	if !ok || len(arr) == 0 {
		return nil
	}
	if r.end == 0 {
		r.end = len(arr)
	}
	if r.start < r.end && r.end <= len(arr) {
		return arr[r.start:r.end]
	}
	return arr
}

func (r Range) String() string {
	return fmt.Sprintf(":range(start: %d, end: %d)", r.start, r.end)
}
