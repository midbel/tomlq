package query

import (
	"reflect"
	"testing"
	"time"
)

var doc = map[string]interface{}{
	"service":   "foobar",
	"instances": []interface{}{1, 2, 3},
	"age":       3600,
	"admin": map[string]interface{}{
		"name":  "midbel",
		"email": "midbel@foobar.org",
		"dob":   time.Date(2020, 10, 12, 14, 0, 0, 0, time.UTC),
	},
	"servers": map[string]interface{}{
		"groups": []interface{}{
			map[string]interface{}{
				"mode":  0,
				"addr":  "239.192.0.1:31001",
				"every": 60,
			},
			map[string]interface{}{
				"mode":  255,
				"addr":  "224.0.0.1:31001",
				"every": 30,
			},
		},
		"prime": map[string]interface{}{
			"addr":   "10.10.1.1:10015",
			"qn":     "prime.foobar.org",
			"reboot": true,
		},
		"backup": map[string]interface{}{
			"addr":   "10.10.1.15:10015",
			"qn":     "backup.foobar.org",
			"reboot": false,
		},
	},
	"client": []interface{}{
		map[string]interface{}{
			"addr": "10.10.0.1:10001",
			"tls":  false,
			"cred": map[string]interface{}{
				"user":   "user1",
				"passwd": "temp123!",
			},
		},
		map[string]interface{}{
			"addr": "10.10.0.2:10001",
			"tls":  true,
			"rps":  50,
			"cred": map[string]interface{}{
				"user":   "user2",
				"passwd": "temp456!",
			},
		},
		map[string]interface{}{
			"addr": "10.10.0.3:10001",
			"tls":  true,
			"rps":  50,
			"cred": map[string]interface{}{
				"user":   "user3",
				"passwd": "temp123!",
			},
		},
	},
}

func TestSelect(t *testing.T) {
	data := []struct {
		Input string
		Want  interface{}
	}{
		{
			Input: ".%service",
			Want:  "foobar",
		},
		{
			Input: "..%service",
			Want:  "foobar",
		},
		{
			Input: ".(service,instances):truthy",
			Want: []interface{}{
				"foobar",
				[]interface{}{1, 2, 3},
			},
		},
		{
			Input: "./[a-z]?e/:number",
			Want:  3600,
		},
		{
			Input: "$admin",
			Want: map[string]interface{}{
				"name":  "midbel",
				"email": "midbel@foobar.org",
				"dob":   time.Date(2020, 10, 12, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			Input: "$admin[email ~= /*@*.org/]",
			Want: map[string]interface{}{
				"name":  "midbel",
				"email": "midbel@foobar.org",
				"dob":   time.Date(2020, 10, 12, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			Input: "$admin[(dob >= 2020-01-01 && dob <= 2020-12-31) || email *= \"foobar\"].%(name,email)",
			Want: []interface{}{
				"midbel",
				"midbel@foobar.org",
			},
		},
		{
			Input: "..$admin[email && (name == \"foobar\" || dob >= 2020-01-01)]",
			Want: map[string]interface{}{
				"name":  "midbel",
				"email": "midbel@foobar.org",
				"dob":   time.Date(2020, 10, 12, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			Input: "..addr",
			Want: []interface{}{
				"239.192.0.1:31001",
				"224.0.0.1:31001",
				"10.10.1.1:10015",
				"10.10.1.15:10015",
				"10.10.0.1:10001",
				"10.10.0.2:10001",
				"10.10.0.3:10001",
			},
		},
		{
			Input: "..@groups[addr ^= \"239\"].%addr:string",
			Want:  "239.192.0.1:31001",
		},
		{
			Input: ".@client[tls == true].addr:truthy",
			Want: []interface{}{
				"10.10.0.2:10001",
				"10.10.0.3:10001",
			},
		},
		{
			Input: ".client[rps].addr",
			Want: []interface{}{
				"10.10.0.2:10001",
				"10.10.0.3:10001",
			},
		},
		{
			Input: "@groups:first",
			Want: map[string]interface{}{
				"mode":  0,
				"addr":  "239.192.0.1:31001",
				"every": 60,
			},
		},
		{
			Input: "@groups:last",
			Want: map[string]interface{}{
				"mode":  255,
				"addr":  "224.0.0.1:31001",
				"every": 30,
			},
		},
	}
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