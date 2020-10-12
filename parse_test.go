package query

import (
	"testing"
)

func TestParse(t *testing.T) {
	data := []struct {
		Input string
		Err   error
	}{
		{
			Input: "foo",
		},
		{
			Input: "/?[a-z]*/",
		},
		{
			Input: "foo.bar",
		},
		{
			Input: "..foo.bar",
		},
		{
			Input: "..$(foo,bar).%bar",
		},
		{
			Input: "..@foo:first",
		},
		{
			Input: "..@foo:at(5)",
		},
		{
			Input: "..@foo:range(0, 10)",
		},
		{
			Input: "..@foo:range(, 10)",
		},
		{
			Input: "..@foo:range(2,)",
		},
		{
			Input: ".foo..bar[str]",
		},
		{
			Input: "..$foo[str == \"value\"].bar",
		},
		{
			Input: "..$foo[str == \"value\"].bar,$foo[int == 0x10].bar",
		},
		{
			Input: "..$foo[str == \"value\" || int == 0x10].bar",
		},
		{
			Input: "foo[str == \"value\" || (int > 0 && int < 9) || pattern ~= /test/]",
		},
		{
			Input: "foo[str && (str^=\"val\" || str$=\"lue\")]",
		},
		{
			Input: "foo[int == (30, 10, 20)]",
		},
	}
	for _, d := range data {
		_, err := Parse(d.Input)
		if err != nil {
			t.Errorf("fail to parse %s: %s", d.Input, err)
			continue
		}
	}
}
