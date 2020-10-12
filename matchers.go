package query

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var ErrNotFound = errors.New("option not found")

type CastError struct {
	value interface{}
	kind  string
}

func (e CastError) Error() string {
	return fmt.Sprintf("%v: fail to cast to %s", e.value, e.kind)
}

func castError(k string, v interface{}) error {
	return CastError{value: v, kind: k}
}

type Infix struct {
	left  Matcher
	right Matcher
	op    rune
}

func (i Infix) String() string {
	var op string
	switch i.op {
	case TokAnd:
		op = "and"
	case TokOr:
		op = "or"
	}
	return fmt.Sprintf("%s(left: %s, right: %s)", op, i.left, i.right)
}

func (i Infix) Match(doc map[string]interface{}) (bool, error) {
	var (
		left  bool
		right bool
		err   error
	)
	if left, err = i.left.Match(doc); err != nil {
		return left, err
	}
	if right, err = i.right.Match(doc); err != nil {
		return right, err
	}
	switch i.op {
	case TokAnd:
		return left && right, nil
	case TokOr:
		return left || right, nil
	default:
		return false, fmt.Errorf("unknown relational operator")
	}
}

type Has struct {
	option string
}

func (h Has) String() string {
	return fmt.Sprintf("has(%s)", h.option)
}

func (h Has) Match(doc map[string]interface{}) (bool, error) {
	_, ok := doc[h.option]
	return ok, nil
}

type Expr struct {
	option string
	value  interface{}
	op     rune
}

func (e Expr) String() string {
	var op string
	switch e.op {
	case TokEqual:
		op = "eq"
	case TokNotEqual:
		op = "ne"
	case TokLesser:
		op = "le"
	case TokLessEq:
		op = "lq"
	case TokGreater:
		op = "gt"
	case TokGreatEq:
		op = "ge"
	case TokStartsWith:
		op = "sw"
	case TokEndsWith:
		op = "ew"
	case TokContains:
		op = "ct"
	case TokMatch:
		op = "mt"
	}
	return fmt.Sprintf("%s(%s, value: %v)", op, e.option, e.value)
}

func (e Expr) Match(doc map[string]interface{}) (bool, error) {
	value, ok := doc[e.option]
	if !ok {
		return ok, fmt.Errorf("%w: %s", ErrNotFound, e.option)
	}

	var err error
	switch es := e.value.(type) {
	case []interface{}:
		for i := range es {
			if ok, err = e.test(es[i], value); ok {
				break
			}
		}
	default:
		ok, err = e.test(e.value, value)
	}
	if err != nil {
		return false, fmt.Errorf("%s: %w", e.option, err)
	}
	return ok, nil
}

func (e Expr) test(want, got interface{}) (bool, error) {
	switch e.op {
	case TokMatch:
		return isMatch(want, got)
	case TokEqual:
		return isEqual(want, got)
	case TokNotEqual:
		ok, err := isEqual(want, got)
		return !ok, err
	case TokLesser:
		return isLess(want, got)
	case TokLessEq:
		eq, err := isEqual(want, got)
		if err != nil {
			return eq, err
		}
		le, err := isLess(want, got)
		if err != nil {
			return le, err
		}
		return eq || le, nil
	case TokGreater:
		eq, err := isEqual(want, got)
		if err != nil {
			return eq, err
		}
		le, err := isLess(want, got)
		if err != nil {
			return le, err
		}
		return !eq && !le, nil
	case TokGreatEq:
		eq, err := isEqual(want, got)
		if err != nil {
			return eq, err
		}
		le, err := isLess(want, got)
		if err != nil {
			return le, err
		}
		return eq || !le, nil
	case TokContains:
		return contains(want, got)
	case TokStartsWith:
		return startsWith(want, got)
	case TokEndsWith:
		return endsWith(want, got)
	default:
	}
	return false, nil
}

func isMatch(want, got interface{}) (bool, error) {
	var (
		pat = want.(string)
		str string
	)
	switch v := got.(type) {
	case int64:
		str = strconv.FormatInt(v, 10)
	case float64:
		str = strconv.FormatFloat(v, 'g', -1, 64)
	case bool:
		str = strconv.FormatBool(v)
	case string:
		str = v
	case time.Time:
		str = v.Format(time.RFC3339)
	default:
		return false, castError("string", got)
	}
	return Match(pat, str), nil
}

func contains(want, got interface{}) (bool, error) {
	val, ok := got.(string)
	if !ok {
		return false, castError("string", got)
	}
	other, ok := want.(string)
	if !ok {
		return false, castError("string", want)
	}
	return strings.Contains(val, other), nil
}

func startsWith(want, got interface{}) (bool, error) {
	val, ok := got.(string)
	if !ok {
		return false, castError("string", got)
	}
	other, ok := want.(string)
	if !ok {
		return false, castError("string", want)
	}
	return strings.HasPrefix(val, other), nil
}

func endsWith(want, got interface{}) (bool, error) {
	val, ok := got.(string)
	if !ok {
		return false, castError("string", got)
	}
	other, ok := want.(string)
	if !ok {
		return false, castError("string", want)
	}
	return strings.HasSuffix(val, other), nil
}

func isEqual(want, got interface{}) (bool, error) {
	switch val := got.(type) {
	case string:
		other, ok := want.(string)
		if !ok {
			return false, castError("string", want)
		}
		return val == other, nil
	case int64:
		other, ok := want.(int64)
		if !ok {
			return false, castError("integer", want)
		}
		return val == other, nil
	case float64:
		other, ok := want.(float64)
		if !ok {
			return false, castError("float", want)
		}
		return val == other, nil
	case bool:
		other, ok := want.(bool)
		if !ok {
			return false, castError("boolean", want)
		}
		return val == other, nil
	case time.Time:
		other, ok := want.(time.Time)
		if !ok {
			return false, castError("time", want)
		}
		return val.Equal(other), nil
	}
	return false, nil
}

func isLess(want, got interface{}) (bool, error) {
	switch val := got.(type) {
	case string:
		other, ok := want.(string)
		if !ok {
			return false, castError("string", want)
		}
		return strings.Compare(other, val) < 0, nil
	case int64:
		other, ok := want.(int64)
		if !ok {
			return false, castError("integer", want)
		}
		return other < val, nil
	case float64:
		other, ok := want.(float64)
		if !ok {
			return false, castError("float", want)
			return false, fmt.Errorf("%v: can not be casted to float", want)
		}
		return other < val, nil
	case bool:
		return false, fmt.Errorf("booleans can only be compared for equality")
	case time.Time:
		other, ok := want.(time.Time)
		if !ok {
			return false, castError("time", want)
		}
		return other.Before(val), nil
	}
	return false, nil
}
