package query

import (
	"reflect"
	"testing"
)

func TestSelector(t *testing.T) {
	data := []struct {
		Data interface{}
		Want interface{}
		Selector
	}{
		{
			Data:     []interface{}{1, 2, 3},
			Want:     1,
			Selector: First{},
		},
		{
			Data:     1,
			Want:     nil,
			Selector: First{},
		},
		{
			Data:     "string",
			Want:     "string",
			Selector: String{},
		},
		{
			Data:     10,
			Want:     nil,
			Selector: String{},
		},
		{
			Data:     int64(0),
			Want:     int64(0),
			Selector: Falsy{},
		},
		{
			Data:     "string",
			Want:     nil,
			Selector: Falsy{},
		},
		{
			Data:     false,
			Want:     false,
			Selector: Boolean{},
		},
		{
			Data:     0.14,
			Want:     nil,
			Selector: Boolean{},
		},
		{
			Data:     0.14,
			Want:     0.14,
			Selector: Float{},
		},
	}
	for _, d := range data {
		got := d.Select(d.Data)
		if !reflect.DeepEqual(d.Want, got) {
			t.Errorf("data mismatched! want %v, got %v", d.Want, got)
		}
	}
}
