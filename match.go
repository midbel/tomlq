package query

import (
	"io"
	"strings"
)

func Match(pattern, input string) bool {
	if pattern == input {
		return true
	}
	var (
		pat = strings.NewReader(pattern)
		str = strings.NewReader(input)
	)
	for pat.Len() > 0 && str.Len() > 0 {
		r, _, _ := pat.ReadRune()
		switch r {
		case star:
			ok := starMatch(pat, str)
			if !ok {
				return ok
			}
		case lsquare:
			ok := rangeMatch(pat, str)
			if !ok {
				return ok
			}
		case question:
			str.ReadRune()
		default:
			x, _, _ := str.ReadRune()
			if x == backslash {
				switch x, _, _ = str.ReadRune(); x {
				case backslash, star, question, lsquare:
				default:
					str.UnreadRune()
				}
			}
			if r != x {
				return false
			}
		}
	}
	return pat.Len() == 0 && str.Len() == 0
}

func rangeMatch(pat, str *strings.Reader) bool {
	accept := func(r rune) bool {
		return isLetter(r) || isDigit(r)
	}
	isRange := func(prev, next rune) bool {
		return prev < next && accept(prev) && accept(next)
	}
	isNegate := func(r rune) bool {
		return r == bang || r == caret
	}

	var (
		negate = true
		prev   rune
		found  bool
	)
	if curr, _, _ := pat.ReadRune(); !isNegate(curr) {
		pat.UnreadRune()
		negate = false
	}
	want, _, _ := str.ReadRune()

	for !found && pat.Len() > 0 {
		curr, _, _ := pat.ReadRune()
		if curr == minus {
			curr, _, _ = pat.ReadRune()
			if !isRange(prev, curr) {
				pat.UnreadRune()
			} else {
				found = want >= prev && want <= curr
				break
			}
		}
		found = curr == want
		prev = curr
	}
	if prev != rsquare {
		for pat.Len() > 0 {
			if curr, _, _ := pat.ReadRune(); curr == rsquare {
				break
			}
		}
	}
	if negate {
		found = !found
	}
	return found
}

func starMatch(pat, str *strings.Reader) bool {
	var next rune
	for pat.Len() > 0 {
		next, _, _ = pat.ReadRune()
		if next != star {
			break
		}
	}
	if pat.Len() == 0 && (next == star || next == 0) {
		str.Seek(0, io.SeekEnd)
		return true
	}
	for str.Len() > 0 {
		if curr, _, _ := str.ReadRune(); curr == next {
			return true
		}
	}
	return false
}
