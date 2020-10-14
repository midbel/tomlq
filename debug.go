package query

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"
)

func Debug(q Queryer, out io.Writer) {
	w := bufio.NewWriter(out)
	defer w.Flush()
	debug(q, w, 0)
}

func debug(q Queryer, out *bufio.Writer, level int) {
	switch qs := q.(type) {
	case Query:
		debugQuery(qs, out, level)
	case Queryset:
		out.WriteString("queryset[\n")
		for _, q := range qs {
			debug(q, out, level+2)
		}
		out.WriteString("]")
	default:
		return
	}
}

func debugQuery(q Query, out *bufio.Writer, level int) {
	write := func(label string, newline bool) {
		space := strings.Repeat(" ", level)
		out.WriteString(space)
		out.WriteString(label)
		if newline {
			out.WriteString("\n")
		}
	}
	writeKV := func(key, val string) {
		space := strings.Repeat(" ", level+2)
		out.WriteString(space)
		out.WriteString(key)
		out.WriteString(val)
		out.WriteString(",\n")
	}
	write("query {", true)
	writeKV("depth  = ", debugDepth(q.depth))
	ks := make([]string, 0, len(q.choices))
	for _, a := range q.choices {
		ks = append(ks, debugAccepter(a))
	}
	writeKV("keys   = ", strings.Join(ks, ", "))
	writeKV("select = ", debugSelector(q.get))

	if q.match != nil {
		writeKV("expr   = ", debugMatcher(q.match))
	}
	if q.next != nil {
		debug(q.next, out, level+2)
	}
	write("}", true)
}

func debugMatcher(m Matcher) string {
	switch e := m.(type) {
	default:
		return "unknown"
	case Expr:
		return debugExpr(e)
	case Infix:
		return debugInfix(e)
	case Has:
		return fmt.Sprintf("exist(%s)", e.option)
	}
	return "matcher"
}

func debugExpr(e Expr) string {
	var op string
	switch e.op {
	case TokEqual:
		op = "eq"
	case TokNotEqual:
		op = "ne"
	case TokLesser:
		op = "ls"
	case TokLessEq:
		op = "le"
	case TokGreater:
		op = "gt"
	case TokGreatEq:
		op = "ge"
	case TokMatch:
		op = "mt"
	case TokContains:
		op = "ct"
	case TokStartsWith:
		op = "sw"
	case TokEndsWith:
		op = "ew"
	}
	valuetype := func(v interface{}) string {
		switch v := v.(type) {
		case int64:
			return fmt.Sprintf("int(%d)", v)
		case float64:
			return fmt.Sprintf("float(%f)", v)
		case bool:
			return fmt.Sprintf("bool(%t)", v)
		case string:
			return fmt.Sprintf("string(%s)", v)
		case time.Time:
			return fmt.Sprintf("datetime(%s)", v.Format(time.RFC3339))
		default:
			return fmt.Sprintf("unknown(%v)", v)
		}
	}
	var vs []string
	switch es := e.value.(type) {
	case []interface{}:
		for _, e := range es {
			vs = append(vs, valuetype(e))
		}
	default:
		vs = append(vs, valuetype(es))
	}
	return fmt.Sprintf("%s(option: %s, values: [%s])", op, e.option, strings.Join(vs, ", "))
}

func debugInfix(e Infix) string {
	var (
		op    string
		left  = debugMatcher(e.left)
		right = debugMatcher(e.right)
	)
	switch e.op {
	case TokAnd:
		op = "and"
	case TokOr:
		op = "or"
	default:
		op = "unknown"
	}
	return fmt.Sprintf("%s(%s, %s)", op, left, right)
}

func debugDepth(depth rune) string {
	switch depth {
	case TokLevelOne:
		return "one"
	case TokLevelAny:
		return "any"
	case TokLevelGreedy:
		return "greedy"
	default:
		return "unknown"
	}
}

func debugAccepter(a Accepter) string {
	var (
		str   string
		label string
		kind  string
		typ   rune
	)
	switch a := a.(type) {
	case Pattern:
		str = "pattern"
		label, typ = a.pattern, a.kind
	case Name:
		str = "label"
		label, typ = a.label, a.kind
	}
	switch typ {
	case TokArray:
		kind = "array"
	case TokRegular:
		kind = "regular"
	case TokValue:
		kind = "value"
	default:
		kind = "any"
	}
	return fmt.Sprintf("%s(%s[%s])", str, label, kind)
}

func debugSelector(sel Selector) string {
	switch s := sel.(type) {
	case At:
		return fmt.Sprintf(":at(index: %d)", s.index)
	case Range:
		return fmt.Sprintf(":range(start: %d, end: %d)", s.start, s.end)
	case First:
		return ":first"
	case Last:
		return ":last"
	case Int:
		return ":int"
	case Float:
		return ":float"
	case Number:
		return ":number"
	case Boolean:
		return ":bool"
	case Truthy:
		return ":truthy"
	case Falsy:
		return ":falsy"
	default:
		return ":all"
	}
}
