package query

import (
	"fmt"
	"strconv"
	"time"
)

type Parser struct {
	scan *Scanner
	curr Token
	peek Token
}

func Parse(str string) (Queryer, error) {
	p := NewParser(str)
	return p.Parse()
}

func NewParser(str string) *Parser {
	var p Parser
	p.scan = NewScanner(str)
	p.next()
	p.next()

	return &p
}

func (p *Parser) Parse() (Queryer, error) {
	return p.parse()
}

func (p *Parser) parse() (Queryer, error) {
	var qs []Queryer
	for !p.isDone() {
		q, err := p.parseQuery()
		if err != nil {
			return nil, err
		}
		qs = append(qs, q)
		switch p.curr.Type {
		case TokComma:
			p.next()
			switch {
			case p.curr.isKey() || p.curr.isLevel() || p.curr.isType():
			default:
				return nil, fmt.Errorf("parse: unexpected token %s", p.curr)
			}
		case TokEOF:
		default:
			return nil, fmt.Errorf("parse: unexpected token %s", p.curr)
		}
	}
	return Queryset(qs), nil
}

func (p *Parser) parseQuery() (Queryer, error) {
	var q Query
	q.depth = TokLevelAny
	if p.curr.isLevel() {
		q.depth = p.curr.Type
		p.next()
	}
	choices, err := p.parseChoices()
	if err != nil {
		return nil, err
	}
	q.choices = choices
	if p.curr.isSelector() {
		get, err := p.parseSelector()
		if err != nil {
			return nil, err
		}
		q.get = get
	}
	if p.curr.isExpression() {
		p.next()
		match, err := p.parseMatcher()
		if err != nil {
			return nil, err
		}
		q.match = match
	}
	if p.curr.isLevel() {
		qs, err := p.parseQuery()
		if err != nil {
			return nil, err
		}
		q.next = qs
	}
	return q, nil
}

func (p *Parser) parseChoices() ([]Accepter, error) {
	var kind rune
	if p.curr.isType() {
		kind = p.curr.Type
		p.next()
	}
	if p.curr.isKey() {
		var a Accepter
		if p.curr.Type == TokPattern {
			a = Pattern{
				pattern: p.curr.Literal,
				kind:    kind,
			}
		} else {
			a = Name{
				label: p.curr.Literal,
				kind:  kind,
			}
		}
		p.next()
		return []Accepter{a}, nil
	}
	if p.curr.Type != TokBegGrp {
		return nil, fmt.Errorf("choices: unexpected token %s, want lparen", p.curr)
	}
	p.next()
	var choices []Accepter
	for !p.isDone() && p.curr.Type != TokEndGrp {
		if !p.curr.isKey() {
			return nil, fmt.Errorf("choices: unexpected token %s, want identifier", p.curr)
		}
		var a Accepter
		if p.curr.Type == TokPattern {
			a = Pattern{
				pattern: p.curr.Literal,
				kind:    kind,
			}
		} else {
			a = Name{
				label: p.curr.Literal,
				kind:  kind,
			}
		}
		choices = append(choices, a)
		p.next()
		switch p.curr.Type {
		case TokComma:
			p.next()
		case TokEndGrp:
		default:
			return nil, fmt.Errorf("choices: unexpected token %s, want comma, rparen", p.curr)
		}
	}
	if p.curr.Type != TokEndGrp {
		return nil, fmt.Errorf("choices: unexpected token %s, want rparen", p.curr)
	}
	p.next()
	return choices, nil
}

func (p *Parser) parseMatcher() (Matcher, error) {
	var left Matcher
	if !p.curr.isKey() {
		return nil, fmt.Errorf("expr: unexpected token %s, want identifier", p.curr)
	}
	ident := p.curr
	p.next()
	if p.curr.isComparison() {
		e := Expr{option: ident.Literal}
		e.op = p.curr.Type
		p.next()
		if !p.curr.isValue() && p.curr.Type != TokBegGrp {
			return nil, fmt.Errorf("expr: unexpected token %s, want value", p.curr)
		}
		value, err := p.parseValue(e.op)
		if err != nil {
			return nil, fmt.Errorf("expr(value): %w", err)
		}
		e.value = value
		left = e
		p.next()
	} else {
		left = Has{option: ident.Literal}
	}
	if p.curr.isRelation() {
		i := Infix{
			op:   p.curr.Type,
			left: left,
		}
		p.next()
		right, err := p.parseMatcher()
		if err != nil {
			return nil, err
		}
		i.right = right
		return i, nil
	}
	if p.curr.Type != TokEndExpr {
		return nil, fmt.Errorf("expr: unexpected token %s, want expr", p.curr)
	}
	p.next()
	return left, nil
}

var timestr = []string{
	"15:04:05",
	"15:04:05.000",
	"15:04:05.000000",
}
var datestr = []string{
	"2006-01-02T15:04:05",
	"2006-01-02T15:04:05.000",
	"2006-01-02T15:04:05.000000",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05.000Z07:00",
	"2006-01-02T15:04:05.000000Z07:00",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04:05.000",
	"2006-01-02 15:04:05.000000",
	"2006-01-02 15:04:05Z07:00",
	"2006-01-02 15:04:05.000Z07:00",
	"2006-01-02 15:04:05.000000Z07:00",
}

func (p *Parser) parseValue(op rune) (interface{}, error) {
	do := func() (interface{}, error) {
		var (
			val interface{}
			err error
		)
		if op == TokMatch && p.curr.Type != TokPattern {
			return nil, fmt.Errorf("value: unexpected token %s, want pattern", p.curr)
		}
		switch p.curr.Type {
		case TokPattern:
			val = p.curr.Literal
		case TokLiteral:
			val = p.curr.Literal
		case TokBool:
			val, err = strconv.ParseBool(p.curr.Literal)
		case TokFloat:
			val, err = strconv.ParseFloat(p.curr.Literal, 64)
		case TokInteger:
			val, err = strconv.ParseInt(p.curr.Literal, 0, 64)
		case TokTime:
			for _, str := range timestr {
				val, err = time.Parse(str, p.curr.Literal)
				if err == nil {
					break
				}
			}
		case TokDate:
			val, err = time.Parse("2006-01-02", p.curr.Literal)
		case TokDateTime:
			for _, str := range datestr {
				val, err = time.Parse(str, p.curr.Literal)
				if err == nil {
					break
				}
			}
		default:
			err = fmt.Errorf("unknown value type: %s", p.curr)
		}
		return val, err
	}
	if p.curr.isValue() {
		val, err := do()
		return []interface{}{val}, err
	}
	if p.curr.Type != TokBegGrp {
		return nil, fmt.Errorf("value: unexpected token %s, want begin", p.curr)
	}
	p.next()

	var values []interface{}
	for !p.isDone() && p.curr.Type != TokEndGrp {
		val, err := do()
		if err != nil {
			return nil, err
		}
		values = append(values, val)
		p.next()
		switch p.curr.Type {
		case TokComma:
			p.next()
		case TokEndGrp:
		default:
			return nil, fmt.Errorf("value: unexpected token %s, want comma|end", p.curr)
		}
	}
	if p.curr.Type != TokEndGrp {
		return nil, fmt.Errorf("value: unexpected token %s, want end", p.curr)
	}
	return values, nil
}

func (p *Parser) parseSelector() (Selector, error) {
	var (
		get Selector
		err error
	)
	switch p.curr.Type {
	case TokSelectAt:
		p.next()
		get, err = p.parseSelectAt()
	case TokSelectFirst:
		p.next()
		get = First{}
	case TokSelectLast:
		p.next()
		get = Last{}
	case TokSelectRange:
		p.next()
		get, err = p.parseSelectRange()
	case TokSelectInt:
		p.next()
		get = Int{}
	case TokSelectFloat:
		p.next()
		get = Float{}
	case TokSelectNumber:
		p.next()
		get = Number{}
	case TokSelectBool:
		p.next()
		get = Boolean{}
	case TokSelectString:
		p.next()
		get = String{}
	case TokSelectTruthy:
		p.next()
		get = Truthy{}
	case TokSelectFalsy:
		p.next()
		get = Falsy{}
	default:
		err = fmt.Errorf("selector: unsupported token %s", p.curr)
	}
	return get, err
}

func (p *Parser) parseSelectAt() (Selector, error) {
	var at At
	if p.curr.Type != TokBegGrp {
		return nil, fmt.Errorf("at: unexpected token %s, want lparen", p.curr)
	}
	p.next()
	ix, err := strconv.ParseInt(p.curr.Literal, 0, 64)
	if err != nil {
		return nil, fmt.Errorf("at: %w", err)
	}
	at.index = int(ix)

	p.next()
	if p.curr.Type != TokEndGrp {
		return nil, fmt.Errorf("at: unexpected token %s, want rparen", p.curr)
	}
	p.next()
	return at, nil
}

func (p *Parser) parseSelectRange() (Selector, error) {
	var rg Range
	if p.curr.Type != TokBegGrp {
		return nil, fmt.Errorf("range: unexpected token %s, want lparen", p.curr)
	}
	p.next()
	if p.curr.Type == TokInteger {
		ix, err := strconv.ParseInt(p.curr.Literal, 0, 64)
		if err != nil {
			return nil, fmt.Errorf("range: %w", err)
		}
		rg.start = int(ix)
		p.next()
	}
	if p.curr.Type != TokComma {
		return nil, fmt.Errorf("range: unexpected token %s, want comma", p.curr)
	}
	p.next()
	if p.curr.Type == TokInteger {
		ix, err := strconv.ParseInt(p.curr.Literal, 0, 64)
		if err != nil {
			return nil, err
		}
		rg.end = int(ix)
		p.next()
	}
	if p.curr.Type != TokEndGrp {
		return nil, fmt.Errorf("range: unexpected token %s, want rparen", p.curr)
	}
	p.next()
	return rg, nil
}

func (p *Parser) isDone() bool {
	return p.curr.isDone()
}

func (p *Parser) next() {
	if p.curr.Type == TokEOF {
		return
	}
	p.curr = p.peek
	p.peek = p.scan.Scan()
}
