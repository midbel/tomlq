package query

import (
	"time"
)

func filterArray(arr []interface{}) interface{} {
	switch len(arr) {
	case 0:
		return nil
	case 1:
		return arr[0]
	default:
		return arr
	}
}

func isArray(v interface{}) bool {
	_, ok := v.([]interface{})
	return ok
}

func isRegular(v interface{}) bool {
	_, ok := v.(map[string]interface{})
	return ok
}

func isValue(v interface{}) bool {
	return !isRegular(v) && !isArray(v)
}

func isTrue(ifi interface{}) bool {
	var ok bool
	switch i := ifi.(type) {
	case int64:
		ok = i != 0
	case float64:
		ok = i != 0
	case bool:
		ok = i == true
	case string:
		ok = len(i) > 0
	case time.Time:
		ok = !i.IsZero()
	case []interface{}:
		ok = len(i) > 0
	case map[string]interface{}:
		ok = len(i) > 0
	default:
		ok = ifi != nil
	}
	return ok
}
