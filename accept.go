package query

import (
	"fmt"
)

type Pattern struct {
	pattern string
	kind    rune
}

func (p Pattern) Accept(ifi map[string]interface{}) (string, interface{}, error) {
	var (
		value interface{}
		key   string
	)
	for k, v := range ifi {
		if Match(p.pattern, k) {
			key, value = k, v
			break
		}
	}
	if value == nil {
		return "", nil, nil
	}
	err := acceptValue(p.kind, value)
	if err != nil {
		err = fmt.Errorf("%s: %w", p.pattern, err)
	}
	return key, value, err
}

func (p Pattern) String() string {
	return p.pattern
}

type Name struct {
	label string
	kind  rune
}

func (n Name) Accept(ifi map[string]interface{}) (string, interface{}, error) {
	value, ok := ifi[n.label]
	if !ok {
		return "", nil, nil
	}
	err := acceptValue(n.kind, value)
	if err != nil {
		err = fmt.Errorf("%s: %w", n.label, err)
	}
	return n.label, value, err
}

func (n Name) String() string {
	return n.label
}

func acceptValue(kind rune, value interface{}) error {
	if kind == 0 {
		return nil
	}
	switch {
	case kind == TokArray && !isArray(value):
		return fmt.Errorf("array expected!")
	case kind == TokRegular && !isRegular(value):
		return fmt.Errorf("table expected!")
	case kind == TokValue && !isValue(value):
		return fmt.Errorf("value expected!")
	default:
		return nil
	}
}
