package query

import (
	"testing"
)

func TestMatch(t *testing.T) {
	data := []struct {
		Pattern string
		Input   string
		Want    bool
	}{
		{
			Pattern: "",
			Input:   "",
			Want:    true,
		},
		{
			Pattern: "*",
			Input:   "",
			Want:    true,
		},
		{
			Pattern: "foobar",
			Input:   "foobar",
			Want:    true,
		},
		{
			Pattern: "foobar",
			Input:   "fOObar",
			Want:    false,
		},
		{
			Pattern: "foo*",
			Input:   "foobar",
			Want:    true,
		},
		{
			Pattern: "foo***",
			Input:   "foobar",
			Want:    true,
		},
		{
			Pattern: "f**bar",
			Input:   "foobar",
			Want:    true,
		},
		{
			Pattern: "f**-bar",
			Input:   "foobar",
			Want:    false,
		},
		{
			Pattern: "f\\o\\obar",
			Input:   "foobar",
			Want:    false,
		},
		{
			Pattern: "f[oO][a-z]???",
			Input:   "foobar",
			Want:    true,
		},
		{
			Pattern: "f[A-Z][a-z]???",
			Input:   "foobar",
			Want:    false,
		},
		{
			Pattern: "f[!A-Z][^a-z]???",
			Input:   "foobar",
			Want:    false,
		},
		{
			Pattern: "f[-0-9]?[]a-z]*",
			Input:   "f--bar",
			Want:    true,
		},
		{
			Pattern: "f\\*\\**",
			Input:   "f**bar",
			Want:    true,
		},
	}
	for _, d := range data {
		got := Match(d.Pattern, d.Input)
		if got != d.Want {
			t.Errorf("%s: match failed %s (want %t, got %t)", d.Input, d.Pattern, d.Want, got)
		}
	}
}
