package query

import (
	"reflect"
	"testing"
	"time"
)

type ParseCase struct {
	Input   string
	Choices []Accepter
	Depth   rune
	Selector
	Matcher
	Next *ParseCase
}

func TestParse(t *testing.T) {
	data := []ParseCase{
		{
			Input: "foo",
			Depth: TokLevelAny,
			Choices: []Accepter{
				createName("foo", 0),
			},
		},
		{
			Input: "/?[a-z]*/",
			Depth: TokLevelAny,
			Choices: []Accepter{
				createPattern("?[a-z]*", 0),
			},
		},
		{
			Input: "foo.bar",
			Depth: TokLevelAny,
			Choices: []Accepter{
				createName("foo", 0),
			},
			Next: &ParseCase{
				Depth: TokLevelOne,
				Choices: []Accepter{
					createName("bar", 0),
				},
			},
		},
		{
			Input: "..foo.(1234, /[a-z][a-z][a-z][a-z]/)",
			Depth: TokLevelAny,
			Choices: []Accepter{
				createName("foo", 0),
			},
			Next: &ParseCase{
				Depth: TokLevelOne,
				Choices: []Accepter{
					createName("1234", 0),
					createPattern("[a-z][a-z][a-z][a-z]", 0),
				},
			},
		},
		{
			Input: "..$(foo,bar).%bar:number",
			Depth: TokLevelAny,
			Choices: []Accepter{
				createName("foo", TokRegular),
				createName("bar", TokRegular),
			},
			Next: &ParseCase{
				Depth: TokLevelOne,
				Choices: []Accepter{
					createName("bar", TokValue),
				},
				Selector: Number{},
			},
		},
		{
			Input: "..@foo:first",
			Depth: TokLevelAny,
			Choices: []Accepter{
				createName("foo", TokArray),
			},
			Selector: First{},
		},
		{
			Input: "..@\"foo\":at(5)",
			Depth: TokLevelAny,
			Choices: []Accepter{
				createName("foo", TokArray),
			},
			Selector: At{5},
		},
		{
			Input: "..@/[a-zA-Z]?*/:range(0, 10)",
			Depth: TokLevelAny,
			Choices: []Accepter{
				createPattern("[a-zA-Z]?*", TokArray),
			},
			Selector: Range{start: 0, end: 10},
		},
		{
			Input: "..@foo:range(, 10)",
			Depth: TokLevelAny,
			Choices: []Accepter{
				createName("foo", TokArray),
			},
			Selector: Range{start: 0, end: 10},
		},
		{
			Input: "..@foo:range(2,)",
			Depth: TokLevelAny,
			Choices: []Accepter{
				createName("foo", TokArray),
			},
			Selector: Range{start: 2, end: 0},
		},
		{
			Input: ".foo..bar[str]",
			Depth: TokLevelOne,
			Choices: []Accepter{
				createName("foo", 0),
			},
			Next: &ParseCase{
				Depth: TokLevelAny,
				Choices: []Accepter{
					createName("bar", 0),
				},
				Matcher: createExist("str"),
			},
		},
		{
			Input: "..$foo[str == \"value\"].bar",
			Depth: TokLevelAny,
			Choices: []Accepter{
				createName("foo", TokRegular),
			},
			Matcher: createExpr(TokEqual, "str", "value"),
			Next: &ParseCase{
				Depth: TokLevelOne,
				Choices: []Accepter{
					createName("bar", 0),
				},
			},
		},
		{
			Input:   "..$foo[str == \"value\"].bar,$foo[int == 0x10].bar",
			Depth:   TokLevelAny,
			Matcher: createExpr(TokEqual, "str", "value"),
			Choices: []Accepter{
				createName("foo", TokRegular),
			},
			Next: &ParseCase{
				Depth: TokLevelOne,
				Choices: []Accepter{
					createName("bar", 0),
				},
			},
		},
		{
			Input: "..$foo[date == 2020-10-12 || time == 13:14:15.678].bar",
			Depth: TokLevelAny,
			Choices: []Accepter{
				createName("foo", TokRegular),
			},
			Matcher: createInfix(TokOr,
				createExpr(TokEqual, "date", time.Date(2020, 10, 12, 0, 0, 0, 0, time.UTC)),
				createExpr(TokEqual, "time", time.Date(0, 1, 1, 13, 14, 15, 678*1000*1000, time.UTC)),
			),
			Next: &ParseCase{
				Depth: TokLevelOne,
				Choices: []Accepter{
					createName("bar", 0),
				},
			},
		},
		{
			Input: "foo[bool == true || (int > 0 && int < 9) || pattern ~= /test/]",
			Depth: TokLevelAny,
			Choices: []Accepter{
				createName("foo", 0),
			},
			Matcher: createInfix(TokOr,
				createExpr(TokEqual, "bool", true),
				createInfix(TokOr,
					createInfix(TokAnd,
						createExpr(TokGreater, "int", int64(0)),
						createExpr(TokLesser, "int", int64(9)),
					),
					createExpr(TokMatch, "pattern", "test"),
				),
			),
		},
		{
			Input: "foo[(int > 0 && int < 9) || (bool == true && pattern ~= /test/)]",
			Depth: TokLevelAny,
			Choices: []Accepter{
				createName("foo", 0),
			},
			Matcher: createInfix(TokOr,
				createInfix(TokAnd,
					createExpr(TokGreater, "int", int64(0)),
					createExpr(TokLesser, "int", int64(9)),
				),
				createInfix(TokAnd,
					createExpr(TokEqual, "bool", true),
					createExpr(TokMatch, "pattern", "test"),
				),
			),
		},
		{
			Input: "foo[str && (str^=\"val\" || str$=\"lue\")]",
			Depth: TokLevelAny,
			Choices: []Accepter{
				createName("foo", 0),
			},
			Matcher: createInfix(TokAnd,
				createExist("str"),
				createInfix(TokOr,
					createExpr(TokStartsWith, "str", "val"),
					createExpr(TokEndsWith, "str", "lue"),
				),
			),
		},
		{
			Input: "foo[int == (30, 10, 20)]",
			Depth: TokLevelAny,
			Choices: []Accepter{
				createName("foo", 0),
			},
			Matcher: createExpr(TokEqual, "int", []interface{}{int64(30), int64(10), int64(20)}),
		},
		{
			Input: "foo[dt == (2020-10-12 13:14:15Z, 2020-10-12T07:08:09.333Z)]",
			Depth: TokLevelAny,
			Choices: []Accepter{
				createName("foo", 0),
			},
			Matcher: createExpr(TokEqual, "dt", []interface{}{
				time.Date(2020, 10, 12, 13, 14, 15, 0, time.UTC),
				time.Date(2020, 10, 12, 7, 8, 9, 333*1000*1000, time.UTC),
			}),
		},
		{
			Input: "foo[pat ~= (/[a-z][0-9]*/, /[A-Z][a-z].???/)]",
			Depth: TokLevelAny,
			Choices: []Accepter{
				createName("foo", 0),
			},
			Matcher: createExpr(TokMatch, "pat", []interface{}{"[a-z][0-9]*", "[A-Z][a-z].???"}),
		},
	}
	for _, d := range data {
		testQuery(t, d)
	}
}

func testQuery(t *testing.T, pc ParseCase) {
	t.Helper()
	q, err := Parse(pc.Input)
	if err != nil {
		t.Errorf("fail to parse %s: %s", pc.Input, err)
		return
	}
	switch qs := q.(type) {
	case Query:
		testSimpleQuery(t, qs, pc)
	case Queryset:
		for _, q := range qs {
			q, ok := q.(Query)
			if !ok {
				t.Errorf("unexpected query type: %T", q)
				return
			}
		}
	default:
		t.Errorf("unexpected query type: %T", q)
		return
	}
}

func testSimpleQuery(t *testing.T, q Query, pc ParseCase) {
	t.Helper()
	if q.depth != pc.Depth {
		t.Errorf("%s: depth mismatched! want %02x, got %02x", pc.Input, pc.Depth, q.depth)
	}
	if !reflect.DeepEqual(q.choices, pc.Choices) {
		t.Errorf("%s: choices mismatched! want %v, got %v", pc.Input, pc.Choices, q.choices)
	}
	if !reflect.DeepEqual(q.get, pc.Selector) {
		t.Errorf("%s: selectors mismatched! want %v, got %v", pc.Input, pc.Selector, q.get)
	}
	if !reflect.DeepEqual(q.match, pc.Matcher) {
		t.Errorf("%s: matchers mismatched!", pc.Input)
		t.Logf("\twant: %v", pc.Matcher)
		t.Logf("\tgot:  %v", q.match)
	}
	if q, ok := q.next.(Query); ok && pc.Next != nil {
		pc.Next.Input = pc.Input
		testSimpleQuery(t, q, *pc.Next)
	}
}

func createExist(str string) Matcher {
	return Has{option: str}
}

func createExpr(op rune, str string, value interface{}) Matcher {
	return Expr{
		option: str,
		value:  value,
		op:     op,
	}
}

func createInfix(op rune, left, right Matcher) Matcher {
	return Infix{
		left:  left,
		right: right,
		op:    op,
	}
}

func createPattern(str string, kind rune) Accepter {
	return Pattern{
		pattern: str,
		kind:    kind,
	}
}

func createName(str string, kind rune) Accepter {
	return Name{
		label: str,
		kind:  kind,
	}
}
