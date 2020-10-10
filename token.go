package query

import (
	"fmt"
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
	var prefix string
	switch t.Type {
	case TokBegExpr, TokBegGrp:
		return "<begin>"
	case TokEndExpr, TokEndGrp:
		return "<end>"
	case TokComma:
		return "<comma>"
	case TokAnd:
		return "<and>"
	case TokOr:
		return "<or>"
	case TokEqual:
		return "<equal>"
	case TokNotEqual:
		return "<notequal>"
	case TokContains:
		return "<contains>"
	case TokMatch:
		return "<match>"
	case TokEndsWith:
		return "<endwidth>"
	case TokStartsWith:
		return "<startwith>"
	case TokLesser:
		return "<less>"
	case TokLessEq:
		return "<lesseq>"
	case TokGreater:
		return "<great>"
	case TokGreatEq:
		return "<greateq>"
	case TokNot:
		return "<not>"
	case TokRegular:
		return "<regular>"
	case TokArray:
		return "<array>"
	case TokValue:
		return "<value>"
	case TokLevelOne:
		return "<one>"
	case TokLevelAny:
		return "<any>"
	case TokEOF:
		return "<eof>"
	case TokLiteral:
		prefix = "literal"
	case TokInteger:
		prefix = "integer"
	case TokFloat:
		prefix = "float"
	case TokBool:
		prefix = "boolean"
	case TokDate:
		prefix = "date"
	case TokTime:
		prefix = "time"
	case TokDateTime:
		prefix = "datetime"
	case TokPattern:
		prefix = "pattern"
	case TokSelectAt:
		return "<:at>"
	case TokSelectRange:
		return "<:range>"
	case TokSelectFirst:
		return "<:first>"
	case TokSelectLast:
		return "<:last>"
	case TokSelectInt:
		return "<:int>"
	case TokSelectFloat:
		return "<:float>"
	case TokSelectNumber:
		return "<:number>"
	case TokSelectBool:
		return "<:bool>"
	case TokSelectString:
		return "<:string>"
	case TokSelectDate:
		return "<:date>"
	case TokSelectTime:
		return "<:time>"
	case TokSelectDatetime:
		return "<:datetime>"
	case TokSelectTruthy:
		return "<:truthy>"
	case TokSelectFalsy:
		return "<:falsy>"
	case TokIllegal:
		prefix = "illegal"
	default:
		prefix = "unknown"
	}
	return fmt.Sprintf("<%s(%s)>", prefix, t.Literal)
}
