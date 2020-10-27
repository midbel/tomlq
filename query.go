package query

import (
	"fmt"
)

type Result struct {
	Paths []string
	Value interface{}
}

func makeResult(ps []string, ifi interface{}) Result {
	return Result {
		Paths: ps,
		Value: ifi,
	}
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
	Select(interface{}) ([]Result, error)
}

type Queryset []Queryer

func (qs Queryset) Select(ifi interface{}) ([]Result, error) {
	rs := make([]Result, 0, len(qs))
	for _, q := range qs {
		r, err := q.Select(ifi)
		if err != nil {
			return nil, err
		}
		rs = append(rs, r...)
	}
	return rs, nil
}

type Query struct {
	choices []Accepter
	depth   rune
	match   Matcher
	get     Selector
	next    Queryer
}

func (q Query) Select(ifi interface{}) ([]Result, error) {
	return q.selectFromInterface(ifi)
}

func (q Query) selectFromInterface(ifi interface{}) ([]Result, error) {
	var (
		err error
		rs  []Result
	)
	switch is := ifi.(type) {
	case []interface{}:
		rs, err = q.selectFromArray(is)
	case map[string]interface{}:
		rs, err = q.selectFromMap(is)
	default:
		return nil, fmt.Errorf("query: can not select from %T", ifi)
	}
	return rs, err
}

func (q Query) selectFromArray(ifi []interface{}) ([]Result, error) {
	var rs []Result
	for _, i := range ifi {
		js, err := q.selectFromInterface(i)
		if err != nil {
			return nil, err
		}
		rs = append(rs, js...)
	}
	return rs, nil
}

func (q Query) selectFromMap(ifi map[string]interface{}) ([]Result, error) {
	rs := make([]Result, 0, len(q.choices))
	for _, key := range q.choices {
		r, err := q.selectFromMapWithKey(key, nil, ifi)
		if err != nil {
			return nil, err
		}
		rs = append(rs, r...)
	}
	return rs, nil
}

func (q Query) selectFromMapWithKey(key Accepter, where []string, ifi map[string]interface{}) ([]Result, error) {
	label, value, err := key.Accept(ifi)
	if err != nil {
		return nil, err
	}
	if q.depth == TokLevelAny && value == nil {
		return q.traverseMap(key, where, ifi)
	}
	if value = q.applySelector(value); value == nil {
		return nil, nil
	}
	rs, err := q.applyMatcher(value, append(where, label))
	if err == nil {
		rs, err = q.applyQuery(rs)
	}
	return rs, err
}

func (q Query) applyQuery(rs []Result) ([]Result, error) {
	if q.next == nil {
		return rs, nil
	}
	xs := make([]Result, 0, len(rs))
	for _, r := range rs {
		if isValue(r.Value) {
			return nil, fmt.Errorf("query: can not apply query to value %v (%q)", r.Value, q.next)
		}
		rs, err := q.next.Select(r.Value)
		if err != nil {
			return nil, err
		}
		for i := range rs {
			rs[i].Paths = append(r.Paths, rs[i].Paths...)
		}
		xs = append(xs, rs...)
	}
	return xs, nil
}

func (q Query) applySelector(ifi interface{}) interface{} {
	if q.get != nil {
		ifi = q.get.Select(ifi)
	}
	return ifi
}

func (q Query) applyMatcher(ifi interface{}, paths []string) ([]Result, error) {
	if q.match == nil {
		return []Result{makeResult(paths, ifi)}, nil
	}
	switch is := ifi.(type) {
	case map[string]interface{}:
		if ok, err := q.match.Match(is); !ok || err != nil {
			return nil, err
		}
		return []Result{makeResult(paths, is)}, nil
	case []interface{}:
		rs := make([]Result, 0, len(is))
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
				rs = append(rs, makeResult(paths, i))
			}
		}
		return rs, nil
	default:
		return nil, fmt.Errorf("query: can not apply predicate to %T", ifi)
	}
}

func (q Query) traverseMap(key Accepter, where []string, ifi map[string]interface{}) ([]Result, error) {
	rs := make([]Result, 0, len(ifi))
	for k, is := range ifi {
		ws := append(where, k)
		switch i := is.(type) {
		case []interface{}:
			vs, err := q.traverseArray(key, ws, i)
			if err != nil {
				return nil, err
			}
			rs = append(rs, vs...)
		case map[string]interface{}:
			vs, err := q.selectFromMapWithKey(key, ws, i)
			if err != nil {
				return nil, err
			}
			rs = append(rs, vs...)
		default:
		}
	}
	return rs, nil
}

func (q Query) traverseArray(key Accepter, where []string, is []interface{}) ([]Result, error) {
	rs := make([]Result, 0, len(is))
	for _, i := range is {
		switch i := i.(type) {
		case map[string]interface{}:
			vs, err := q.selectFromMapWithKey(key, where, i)
			if err != nil {
				return nil, err
			}
			rs = append(rs, vs...)
		case []interface{}:
			vs, err := q.traverseArray(key, where, i)
			if err != nil {
				return nil, err
			}
			rs = append(rs, vs...)
		}
	}
	return rs, nil
}
