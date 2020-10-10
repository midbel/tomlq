package query

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

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
		return ok, fmt.Errorf("%s: option not found", e.option)
	}
	switch e.op {
	case TokMatch:
		return e.match(value)
	case TokEqual:
		return e.isEqual(value)
	case TokNotEqual:
		ok, err := e.isEqual(value)
		return !ok, err
	case TokLesser:
		return e.isLess(value)
	case TokLessEq:
		eq, err := e.isEqual(value)
		if err != nil {
			return eq, err
		}
		le, err := e.isLess(value)
		if err != nil {
			return le, err
		}
		return eq || le, nil
	case TokGreater:
		eq, err := e.isEqual(value)
		if err != nil {
			return eq, err
		}
		le, err := e.isLess(value)
		if err != nil {
			return le, err
		}
		return !eq && !le, nil
	case TokGreatEq:
		eq, err := e.isEqual(value)
		if err != nil {
			return eq, err
		}
		le, err := e.isLess(value)
		if err != nil {
			return le, err
		}
		return eq || !le, nil
	case TokContains:
		return e.contains(value)
	case TokStartsWith:
		return e.startsWith(value)
	case TokEndsWith:
		return e.endsWith(value)
	default:
	}
	return false, nil
}

func (e Expr) match(value interface{}) (bool, error) {
	var (
		pat = e.value.(string)
		str string
	)
	switch v := value.(type) {
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
		return false, fmt.Errorf("%v: can not be converted to string", value)
	}
	return Match(pat, str), nil
}

func (e Expr) contains(value interface{}) (bool, error) {
	val, ok := e.value.(string)
	if !ok {
		return false, fmt.Errorf("%s: option can not be cast to string", e.option)
	}
	other, ok := value.(string)
	if !ok {
		return false, fmt.Errorf("%v: can not be cast to string", value)
	}
	return strings.Contains(val, other), nil
}

func (e Expr) startsWith(value interface{}) (bool, error) {
	val, ok := e.value.(string)
	if !ok {
		return false, fmt.Errorf("%s: option can not be cast to string", e.option)
	}
	other, ok := value.(string)
	if !ok {
		return false, fmt.Errorf("%v: can not be cast to string", value)
	}
	return strings.HasPrefix(val, other), nil
}

func (e Expr) endsWith(value interface{}) (bool, error) {
	val, ok := e.value.(string)
	if !ok {
		return false, fmt.Errorf("%s: option can not be cast to string", e.option)
	}
	other, ok := value.(string)
	if !ok {
		return false, fmt.Errorf("%v: can not be cast to string", value)
	}
	return strings.HasSuffix(val, other), nil
}

func (e Expr) isEqual(value interface{}) (bool, error) {
	switch val := e.value.(type) {
	case string:
		other, ok := value.(string)
		if !ok {
			return false, fmt.Errorf("%s(%v): can not be casted to string", e.option, value)
		}
		return val == other, nil
	case int64:
		other, ok := value.(int64)
		if !ok {
			return false, fmt.Errorf("%s(%v): can not be casted to integer", e.option, value)
		}
		return val == other, nil
	case float64:
		other, ok := value.(float64)
		if !ok {
			return false, fmt.Errorf("%s(%v): can not be casted to float", e.option, value)
		}
		return val == other, nil
	case bool:
		other, ok := value.(bool)
		if !ok {
			return false, fmt.Errorf("%s(%v): can not be casted to boolean", e.option, value)
		}
		return val == other, nil
	case time.Time:
		other, ok := value.(time.Time)
		if !ok {
			return false, fmt.Errorf("%s(%v): can not be casted to time", e.option, value)
		}
		return val.Equal(other), nil
	}
	return false, nil
}

func (e Expr) isLess(value interface{}) (bool, error) {
	switch val := e.value.(type) {
	case string:
		other, ok := value.(string)
		if !ok {
			return false, fmt.Errorf("%s(%v): can not be casted to string", e.option, value)
		}
		return strings.Compare(other, val) < 0, nil
	case int64:
		other, ok := value.(int64)
		if !ok {
			return false, fmt.Errorf("%s(%v): can not be casted to integer", e.option, value)
		}
		return other < val, nil
	case float64:
		other, ok := value.(float64)
		if !ok {
			return false, fmt.Errorf("%s(%v): can not be casted to float", e.option, value)
		}
		return other < val, nil
	case bool:
		return false, fmt.Errorf("booleans can only be compared for equality")
	case time.Time:
		other, ok := value.(time.Time)
		if !ok {
			return false, fmt.Errorf("%s(%v): can not be casted to time", e.option, value)
		}
		return other.Before(val), nil
	}
	return false, nil
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
