package query

import (
	"reflect"
	"testing"
)

var doc = map[string]interface{}{}

func TestSelect(t *testing.T) {
	data := []struct {
		Input string
		Want  interface{}
	}{}
	for _, d := range data {
		q, err := Parse(d.Input)
		if err != nil {
			t.Errorf("error parsing %s: %s", d.Input, err)
			continue
		}
		got, err := q.Select(doc)
		if err != nil {
			t.Errorf("error fetching data: %s", err)
			continue
		}
		if !reflect.DeepEqual(d.Want, got) {
			t.Errorf("%s: results mismatched!", d.Input)
			t.Logf("\twant: %v", d.Want)
			t.Logf("\tgot:  %v", got)
		}
	}
}
