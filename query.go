package query

import (
	"fmt"
	"strings"
)

type Value struct {
	Paths []string
	Label string
	Value interface{}
}

type Selector interface {
	Select(interface{}) interface{}
}

type Matcher interface {
	Match(map[string]interface{}) (bool, error)
}

type Accepter interface {
	Accept(map[string]interface{}) (string, interface{}, error)
	fmt.Stringer
}

type Queryer interface {
	Select(interface{}) (interface{}, error)
}

type Queryset []Queryer

func (qs Queryset) Select(ifi interface{}) (interface{}, error) {
	is := make([]interface{}, 0, len(qs))
	for _, q := range qs {
		i, err := q.Select(ifi)
		if err != nil {
			return nil, err
		}
		is = append(is, i)
	}
	return filterArray(is), nil
}

type Query struct {
	choices []Accepter
	depth rune
	match Matcher
	get   Selector
	next  Queryer
}

func (q Query) Select(ifi interface{}) (interface{}, error) {
	return q.selectFromInterface(ifi)
}

func (q Query) selectFromInterface(ifi interface{}) (interface{}, error) {
	var err error
	switch is := ifi.(type) {
	case []interface{}:
		ifi, err = q.selectFromArray(is)
	case map[string]interface{}:
		ifi, err = q.selectFromMap(is)
	default:
	}
	return ifi, err
}

func (q Query) selectFromArray(ifi []interface{}) (interface{}, error) {
	var is []interface{}
	for _, i := range ifi {
		j, err := q.selectFromInterface(i)
		if err != nil {
			return nil, err
		}
		if j != nil {
			is = append(is, j)
		}
	}
	return filterArray(is), nil
}

func (q Query) selectFromMap(ifi map[string]interface{}) (interface{}, error) {
	is := make([]interface{}, 0, len(q.choices))
	for _, key := range q.choices {
		i, err := q.selectFromMapWithKey(key, ifi)
		if err != nil {
			return nil, err
		}
		if i != nil {
			is = append(is, i)
		}
	}
	return filterArray(is), nil
}

func (q Query) selectFromMapWithKey(key Accepter, ifi map[string]interface{}) (interface{}, error) {
	_, value, err := key.Accept(ifi)
	if err != nil {
		return nil, err
	}
	if q.depth == TokLevelAny && value == nil {
		return q.traverseMap(key, ifi)
	}
	if value = q.applySelector(value); value == nil {
		return nil, nil
	}
	if value, err = q.applyMatcher(value); value == nil || err != nil {
		return value, err
	}
	return q.applyQuery(value)
}

func (q Query) traverseMap(key Accepter, ifi map[string]interface{}) (interface{}, error) {
	qs := make([]interface{}, 0, len(ifi))
	for _, is := range ifi {
		switch i := is.(type) {
		case []interface{}:
			vs, err := q.traverseArray(key, i)
			if err != nil {
				return nil, err
			}
			if len(vs) > 0 {
				qs = append(qs, vs...)
			}
		case map[string]interface{}:
			v, err := q.selectFromMapWithKey(key, i)
			if err != nil {
				return nil, err
			}
			if v != nil {
				qs = append(qs, v)
			}
		default:
		}
	}
	return filterArray(qs), nil
}

func (q Query) traverseArray(key Accepter, is []interface{}) ([]interface{}, error) {
	isEmpty := func(v interface{}) bool {
		if v == nil {
			return true
		}
		switch v := v.(type) {
		default:
			return false
		case []interface{}:
			return len(v) == 0
		case map[string]interface{}:
			return len(v) == 0
		}
	}
	qs := make([]interface{}, 0, len(is))
	for _, i := range is {
		switch i := i.(type) {
		case map[string]interface{}:
			v, err := q.selectFromMapWithKey(key, i)
			if err != nil {
				return nil, err
			}
			if !isEmpty(v) {
				qs = append(qs, v)
			}
		case []interface{}:
			vs, err := q.traverseArray(key, i)
			if err != nil {
				return nil, err
			}
			if len(vs) > 0 {
				qs = append(qs, vs...)
			}
		}
	}
	return qs, nil
}

func (q Query) applyQuery(ifi interface{}) (interface{}, error) {
	if q.next == nil {
		return ifi, nil
	}
	if isValue(ifi) {
		return nil, fmt.Errorf("query: can not apply query to value %v (%q)", ifi, q.next)
	}
	return q.next.Select(ifi)
}

func (q Query) applySelector(ifi interface{}) interface{} {
	if q.get != nil {
		ifi = q.get.Select(ifi)
	}
	return ifi
}

func (q Query) applyMatcher(ifi interface{}) (interface{}, error) {
	if q.match == nil {
		return ifi, nil
	}
	if isValue(ifi) {
		return nil, fmt.Errorf("match: can not apply predicate to value %v (%s)", ifi, q.match)
	}
	switch is := ifi.(type) {
	case map[string]interface{}:
		if ok, err := q.match.Match(is); !ok || err != nil {
			return nil, err
		}
	case []interface{}:
		xs := make([]interface{}, 0, len(is))
		for _, i := range is {
			i, ok := i.(map[string]interface{})
			if !ok {
				continue
			}
			ok, err := q.match.Match(i)
			if err != nil {
				return nil, err
			}
			if ok {
				xs = append(xs, i)
			}
		}
		ifi = xs
	default:
	}
	return ifi, nil
}

const qPattern = "query(%s, depth: %s, select: %v, match: %v, next: %v)"

func (q Query) String() string {
	var depth string
	switch q.depth {
	case TokLevelOne:
		depth = "current"
	default:
		depth = "any"
	}
	cs := make([]string, len(q.choices))
	for _, c := range q.choices {
		cs = append(cs, c.String())
	}
	return fmt.Sprintf(qPattern, strings.Join(cs, "|"), depth, q.get, q.match, q.next)
}
