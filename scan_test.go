package query

import (
	"testing"
)

func TestScanner(t *testing.T) {
	data := []struct {
		Input  string
		Tokens []Token
	}{
		{
			Input: "foo",
			Tokens: []Token{
				createToken("foo", TokLiteral),
			},
		},
		{
			Input: "..foo,.bar",
			Tokens: []Token{
				createToken("", TokLevelAny),
				createToken("foo", TokLiteral),
				createToken("", TokComma),
				createToken("", TokLevelOne),
				createToken("bar", TokLiteral),
			},
		},
		{
			Input: ".foo",
			Tokens: []Token{
				createToken("", TokLevelOne),
				createToken("foo", TokLiteral),
			},
		},
		{
			Input: ".foo:int",
			Tokens: []Token{
				createToken("", TokLevelOne),
				createToken("foo", TokLiteral),
				createToken("int", TokSelectInt),
			},
		},
		{
			Input: ".foo:at(1)",
			Tokens: []Token{
				createToken("", TokLevelOne),
				createToken("foo", TokLiteral),
				createToken("at", TokSelectAt),
				createToken("", TokBegGrp),
				createToken("1", TokInteger),
				createToken("", TokEndGrp),
			},
		},
		{
			Input: ".foo..$1234",
			Tokens: []Token{
				createToken("", TokLevelOne),
				createToken("foo", TokLiteral),
				createToken("", TokLevelAny),
				createToken("", TokRegular),
				createToken("1234", TokInteger),
			},
		},
		{
			Input: ".foo..$\"bar\"",
			Tokens: []Token{
				createToken("", TokLevelOne),
				createToken("foo", TokLiteral),
				createToken("", TokLevelAny),
				createToken("", TokRegular),
				createToken("bar", TokLiteral),
			},
		},
		{
			Input: ".foo..%/[a-z]?*/",
			Tokens: []Token{
				createToken("", TokLevelOne),
				createToken("foo", TokLiteral),
				createToken("", TokLevelAny),
				createToken("", TokValue),
				createToken("[a-z]?*", TokPattern),
			},
		},
		{
			Input: ".(foo,bar)",
			Tokens: []Token{
				createToken("", TokLevelOne),
				createToken("", TokBegGrp),
				createToken("foo", TokLiteral),
				createToken("", TokComma),
				createToken("bar", TokLiteral),
				createToken("", TokEndGrp),
			},
		},
		{
			Input: "..@(foo,bar)",
			Tokens: []Token{
				createToken("", TokLevelAny),
				createToken("", TokArray),
				createToken("", TokBegGrp),
				createToken("foo", TokLiteral),
				createToken("", TokComma),
				createToken("bar", TokLiteral),
				createToken("", TokEndGrp),
			},
		},
		{
			Input: "foo[bar == \"value\"]",
			Tokens: []Token{
				createToken("foo", TokLiteral),
				createToken("", TokBegExpr),
				createToken("bar", TokLiteral),
				createToken("", TokEqual),
				createToken("value", TokLiteral),
				createToken("", TokEndExpr),
			},
		},
		{
			Input: "foo[bar $= \"value\"]",
			Tokens: []Token{
				createToken("foo", TokLiteral),
				createToken("", TokBegExpr),
				createToken("bar", TokLiteral),
				createToken("", TokEndsWith),
				createToken("value", TokLiteral),
				createToken("", TokEndExpr),
			},
		},
		{
			Input: "foo[bar ^= \"value\"]",
			Tokens: []Token{
				createToken("foo", TokLiteral),
				createToken("", TokBegExpr),
				createToken("bar", TokLiteral),
				createToken("", TokStartsWith),
				createToken("value", TokLiteral),
				createToken("", TokEndExpr),
			},
		},
		{
			Input: "foo[bar *= \"value\"]",
			Tokens: []Token{
				createToken("foo", TokLiteral),
				createToken("", TokBegExpr),
				createToken("bar", TokLiteral),
				createToken("", TokContains),
				createToken("value", TokLiteral),
				createToken("", TokEndExpr),
			},
		},
		{
			Input: "foo[bar ~= /[a-z]*?/]",
			Tokens: []Token{
				createToken("foo", TokLiteral),
				createToken("", TokBegExpr),
				createToken("bar", TokLiteral),
				createToken("", TokMatch),
				createToken("[a-z]*?", TokPattern),
				createToken("", TokEndExpr),
			},
		},
		{
			Input: "foo[bar != true]",
			Tokens: []Token{
				createToken("foo", TokLiteral),
				createToken("", TokBegExpr),
				createToken("bar", TokLiteral),
				createToken("", TokNotEqual),
				createToken("true", TokBool),
				createToken("", TokEndExpr),
			},
		},
		{
			Input: "foo[bar <= 0xca_fe]",
			Tokens: []Token{
				createToken("foo", TokLiteral),
				createToken("", TokBegExpr),
				createToken("bar", TokLiteral),
				createToken("", TokLessEq),
				createToken("0xcafe", TokInteger),
				createToken("", TokEndExpr),
			},
		},
		{
			Input: "foo[bar <= 0b1_1_1_1]",
			Tokens: []Token{
				createToken("foo", TokLiteral),
				createToken("", TokBegExpr),
				createToken("bar", TokLiteral),
				createToken("", TokLessEq),
				createToken("0b1111", TokInteger),
				createToken("", TokEndExpr),
			},
		},
		{
			Input: "foo[bar <= 0o45_67]",
			Tokens: []Token{
				createToken("foo", TokLiteral),
				createToken("", TokBegExpr),
				createToken("bar", TokLiteral),
				createToken("", TokLessEq),
				createToken("0o4567", TokInteger),
				createToken("", TokEndExpr),
			},
		},
		{
			Input: "foo[bar <= 123_456]",
			Tokens: []Token{
				createToken("foo", TokLiteral),
				createToken("", TokBegExpr),
				createToken("bar", TokLiteral),
				createToken("", TokLessEq),
				createToken("123_456", TokInteger),
				createToken("", TokEndExpr),
			},
		},
		{
			Input: "foo[bar > -0.14e+4]",
			Tokens: []Token{
				createToken("foo", TokLiteral),
				createToken("", TokBegExpr),
				createToken("bar", TokLiteral),
				createToken("", TokGreater),
				createToken("-0.14e+4", TokFloat),
				createToken("", TokEndExpr),
			},
		},
		{
			Input: "foo[bar >= 2020-10-12]",
			Tokens: []Token{
				createToken("foo", TokLiteral),
				createToken("", TokBegExpr),
				createToken("bar", TokLiteral),
				createToken("", TokGreatEq),
				createToken("2020-10-12", TokDate),
				createToken("", TokEndExpr),
			},
		},
		{
			Input: "foo[bar == 2020-10-12 10:20:30.789+02:00]",
			Tokens: []Token{
				createToken("foo", TokLiteral),
				createToken("", TokBegExpr),
				createToken("bar", TokLiteral),
				createToken("", TokEqual),
				createToken("2020-10-12 10:20:30.789+02:00", TokDateTime),
				createToken("", TokEndExpr),
			},
		},
		{
			Input: "foo[bar < 10:20:30.789]",
			Tokens: []Token{
				createToken("foo", TokLiteral),
				createToken("", TokBegExpr),
				createToken("bar", TokLiteral),
				createToken("", TokLesser),
				createToken("10:20:30.789", TokTime),
				createToken("", TokEndExpr),
			},
		},
		{
			Input: "foo[str == \"value\" && float == 0.123_456 || date != 2020-10-12]",
			Tokens: []Token{
				createToken("foo", TokLiteral),
				createToken("", TokBegExpr),
				createToken("str", TokLiteral),
				createToken("", TokEqual),
				createToken("value", TokLiteral),
				createToken("", TokAnd),
				createToken("float", TokLiteral),
				createToken("", TokEqual),
				createToken("0.123_456", TokFloat),
				createToken("", TokOr),
				createToken("date", TokLiteral),
				createToken("", TokNotEqual),
				createToken("2020-10-12", TokDate),
				createToken("", TokEndExpr),
			},
		},
		{
			Input: "foo[int == (10, 0, 20)]",
			Tokens: []Token{
				createToken("foo", TokLiteral),
				createToken("", TokBegExpr),
				createToken("int", TokLiteral),
				createToken("", TokEqual),
				createToken("", TokBegGrp),
				createToken("10", TokInteger),
				createToken("", TokComma),
				createToken("0", TokInteger),
				createToken("", TokComma),
				createToken("20", TokInteger),
				createToken("", TokEndGrp),
				createToken("", TokEndExpr),
			},
		},
	}
	for _, d := range data {
		s := NewScanner(d.Input)
		for i := 0; ; i++ {
			got := s.Scan()
			if got.Type == TokEOF {
				break
			}
			if i >= len(d.Tokens) {
				t.Errorf("too many tokens created! got %s (%d - want %d)", got, i, len(d.Tokens))
				break
			}
			if want := d.Tokens[i]; !compareTokens(want, got) {
				t.Errorf("%d: tokens mismatched! want %s, got %s", i, want, got)
				break
			}
		}
	}
}

func compareTokens(fst, snd Token) bool {
	return fst.Literal == snd.Literal && fst.Type == snd.Type
}

func createToken(str string, kind rune) Token {
	return Token{
		Literal: str,
		Type:    kind,
	}
}
