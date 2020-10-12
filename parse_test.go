package query

import (
	"testing"
)

func TestParse(t *testing.T) {
	data := []string{
		"foo",
		"/?[a-z]*/",
		"foo.bar",
		"..foo.bar",
		"..$(foo,bar).%bar",
		"..@foo:first",
		"..@\"foo\":at(5)",
		"..@/[a-zA-Z]?*/:range(0, 10)",
		"..@foo:range(, 10)",
		"..@foo:range(2,)",
		".foo..bar[str]",
		"..$foo[str == \"value\"].bar",
		"..$foo[str == \"value\"].bar,$foo[int == 0x10].bar",
		"..$foo[str == \"value\" || int == 0x10].bar",
		"foo[bool == true || (int > 0 && int < 9) || pattern ~= /test/]",
		"foo[(int > 0 && int < 9) || (bool == true && pattern ~= /test/)]",
		"foo[str && (str^=\"val\" || str$=\"lue\")]",
		"foo[int == (30, 10, 20)]",
	}
	for _, d := range data {
		_, err := Parse(d)
		if err != nil {
			t.Errorf("fail to parse %s: %s", d, err)
			continue
		}
	}
}
