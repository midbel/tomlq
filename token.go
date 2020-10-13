package query

import (
	"fmt"
	"sort"
)

const (
	TokEOF rune = -(iota + 1)
	TokLiteral
	TokInteger
	TokFloat
	TokBool
	TokTime
	TokDate
	TokDateTime
	TokPattern
	TokIllegal
	TokLevelOne
	TokLevelAny
	TokArray
	TokRegular
	TokValue
	TokEqual
	TokNotEqual
	TokLesser
	TokLessEq
	TokGreater
	TokGreatEq
	TokStartsWith
	TokEndsWith
	TokContains
	TokMatch
	TokAnd
	TokOr
	TokBegExpr
	TokEndExpr
	TokBegGrp
	TokEndGrp
	TokComma
	TokNot
	TokSelectAt
	TokSelectRange
	TokSelectFirst
	TokSelectLast
	TokSelectInt
	TokSelectFloat
	TokSelectNumber
	TokSelectBool
	TokSelectString
	TokSelectDate
	TokSelectTime
	TokSelectDatetime
	TokSelectTruthy
	TokSelectFalsy
)

var identifiers = map[string]rune{
	"true":  TokBool,
	"false": TokBool,
}

var selectors = map[string]rune{
	"first":    TokSelectFirst,
	"last":     TokSelectLast,
	"range":    TokSelectRange,
	"at":       TokSelectAt,
	"int":      TokSelectInt,
	"float":    TokSelectFloat,
	"number":   TokSelectNumber,
	"bool":     TokSelectBool,
	"string":   TokSelectString,
	"date":     TokSelectDate,
	"time":     TokSelectTime,
	"datetime": TokSelectDatetime,
	"truthy":   TokSelectTruthy,
	"falsy":    TokSelectFalsy,
}

var typenames = []struct {
	Label    string
	Type     rune
	Compound bool
}{
	{Label: "literal", Type: TokLiteral, Compound: true},
	{Label: "integer", Type: TokInteger, Compound: true},
	{Label: "float", Type: TokFloat, Compound: true},
	{Label: "boolean", Type: TokBool, Compound: true},
	{Label: "date", Type: TokDate, Compound: true},
	{Label: "time", Type: TokTime, Compound: true},
	{Label: "datetime", Type: TokDateTime, Compound: true},
	{Label: "pattern", Type: TokPattern, Compound: true},
	{Label: "illegal", Type: TokIllegal, Compound: true},
	{Label: ":at", Type: TokSelectAt},
	{Label: ":range", Type: TokSelectRange},
	{Label: ":first", Type: TokSelectFirst},
	{Label: ":last", Type: TokSelectLast},
	{Label: ":int", Type: TokSelectInt},
	{Label: ":float", Type: TokSelectFloat},
	{Label: ":number", Type: TokSelectNumber},
	{Label: ":bool", Type: TokSelectBool},
	{Label: ":string", Type: TokSelectString},
	{Label: ":date", Type: TokSelectDate},
	{Label: ":time", Type: TokSelectTime},
	{Label: ":datetime", Type: TokSelectDatetime},
	{Label: ":truthy", Type: TokSelectTruthy},
	{Label: ":falsy", Type: TokSelectFalsy},
	{Label: "comma", Type: TokComma},
	{Label: "and", Type: TokAnd},
	{Label: "or", Type: TokOr},
	{Label: "equal", Type: TokEqual},
	{Label: "notequal", Type: TokNotEqual},
	{Label: "contains", Type: TokContains},
	{Label: "match", Type: TokMatch},
	{Label: "endswith", Type: TokEndsWith},
	{Label: "startswith", Type: TokStartsWith},
	{Label: "lesser", Type: TokLesser},
	{Label: "lesseq", Type: TokLessEq},
	{Label: "greater", Type: TokGreater},
	{Label: "greateq", Type: TokGreatEq},
	{Label: "not", Type: TokNot},
	{Label: "regular", Type: TokRegular},
	{Label: "array", Type: TokArray},
	{Label: "value", Type: TokValue},
	{Label: "one", Type: TokLevelOne},
	{Label: "any", Type: TokLevelAny},
	{Label: "eof", Type: TokEOF},
	{Label: "beg-expr", Type: TokBegExpr},
	{Label: "end-expr", Type: TokEndExpr},
	{Label: "beg-grp", Type: TokBegGrp},
	{Label: "end-grp", Type: TokEndGrp},
}

func init() {
	sort.Slice(typenames, func(i, j int) bool {
		return typenames[i].Type <= typenames[j].Type
	})
}

type Token struct {
	Literal string
	Type    rune
}

func (t Token) isValue() bool {
	switch t.Type {
	case TokLiteral, TokPattern, TokBool, TokInteger, TokFloat:
	case TokTime, TokDate, TokDateTime:
	default:
		return false
	}
	return true
}

func (t Token) isKey() bool {
	return t.Type == TokLiteral || t.Type == TokInteger || t.Type == TokPattern
}

func (t Token) isType() bool {
	return t.Type == TokValue || t.Type == TokArray || t.Type == TokRegular
}

func (t Token) isLevel() bool {
	return t.Type == TokLevelOne || t.Type == TokLevelAny
}

func (t Token) isSelector() bool {
	switch t.Type {
	case TokSelectAt, TokSelectRange, TokSelectFirst, TokSelectLast:
	case TokSelectInt, TokSelectFloat, TokSelectNumber, TokSelectBool, TokSelectString:
	case TokSelectTruthy, TokSelectFalsy:
	default:
		return false
	}
	return true
}

func (t Token) isComparison() bool {
	switch t.Type {
	case TokEqual, TokNotEqual, TokContains, TokStartsWith, TokEndsWith, TokMatch:
	case TokLesser, TokGreater, TokLessEq, TokGreatEq:
	default:
		return false
	}
	return true
}

func (t Token) isRelation() bool {
	return t.Type == TokAnd || t.Type == TokOr
}

func (t Token) isExpression() bool {
	return t.Type == TokBegExpr
}

func (t Token) isDone() bool {
	return t.Type == TokEOF || t.Type == TokIllegal
}

func (t Token) String() string {
	x := sort.Search(len(typenames), func(i int) bool {
		return t.Type <= typenames[i].Type
	})
	if x < len(typenames) && typenames[x].Type == t.Type {
		str := typenames[x]
		if str.Compound {
			return fmt.Sprintf("<%s(%s)>", str.Label, t.Literal)
		}
		return fmt.Sprintf("<%s>", str.Label)
	}
	return fmt.Sprintf("<unknown(%s)>", t.Literal)
}
