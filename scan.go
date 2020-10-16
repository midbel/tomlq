package query

import (
	"bytes"
	"unicode/utf8"
)

const (
	squote     = '\''
	dquote     = '"'
	dot        = '.'
	minus      = '-'
	underscore = '_'
	percent    = '%'
	dollar     = '$'
	arobase    = '@'
	lsquare    = '['
	rsquare    = ']'
	equal      = '='
	bang       = '!'
	langle     = '<'
	rangle     = '>'
	ampersand  = '&'
	pipe       = '|'
	space      = ' '
	tab        = '\t'
	colon      = ':'
	lparen     = '('
	rparen     = ')'
	comma      = ','
	caret      = '^'
	plus       = '+'
	question   = '?'
	star       = '*'
	slash      = '/'
	tilde      = '~'
	backslash  = '\\'
	formfeed   = '\f'
	backspace  = '\b'
	carriage   = '\r'
	newline    = '\n'
	zero       = '0'
	hex        = 'x'
	oct        = 'o'
	bin        = 'b'
)

var escapes = map[rune]rune{
	backslash: backslash,
	dquote:    dquote,
	'n':       newline,
	't':       tab,
	'f':       formfeed,
	'b':       backspace,
	'r':       carriage,
}

var constants = map[string]rune{
	"true":  TokBool,
	"false": TokBool,
	"inf":   TokFloat,
	"+inf":  TokFloat,
	"-inf":  TokFloat,
	"nan":   TokFloat,
	"+nan":  TokFloat,
	"-nan":  TokFloat,
}

type Scanner struct {
	buffer []byte
	char   rune
	curr   int
	next   int

	expr bool
	scan func() Token
}

func NewScanner(str string) *Scanner {
	var s Scanner
	s.buffer = []byte(str)
	s.scan = s.scanDefault

	s.readRune()
	return &s
}

func (s *Scanner) Scan() Token {
	tok := s.scan()
	switch tok.Type {
	case TokBegExpr:
		s.scan = s.scanExpr
	case TokEndExpr:
		s.scan = s.scanDefault
	}
	return tok
}

func (s *Scanner) scanExpr() Token {
	var tok Token
	if s.isDone() {
		tok.Type = TokIllegal
		return tok
	}
	s.skip(isBlank)
	pos := s.curr
	switch {
	case isQuote(s.char):
		s.scanQuote(&tok)
	case isOperator(s.char):
		s.scanOperator(&tok)
	case isDigit(s.char) || isSign(s.char):
		k := s.nextRune()
		if s.char == zero && (k == hex || k == bin || k == oct) {
			s.scanBase(&tok)
		} else {
			s.scanNumber(&tok)
		}
	case isAlpha(s.char):
		s.scanLiteral(&tok)
	case isControl(s.char):
		s.scanControl(&tok)
	case isPattern(s.char):
		s.scanPattern(&tok)
	default:
		tok.Type = TokIllegal
	}
	if tok.Type == TokIllegal && s.curr > pos {
		s.reset(pos)
		tok.Literal = s.scanIllegal(func(r rune) bool { return isControl(r) || isOperator(r) })
	}
	return tok
}

func (s *Scanner) scanDefault() Token {
	var tok Token
	if s.isDone() {
		tok.Type = s.char
		return tok
	}
	pos := s.curr
	switch {
	case isDigit(s.char):
		s.scanDigit(&tok)
	case isLetter(s.char):
		s.scanLiteral(&tok)
	case isQuote(s.char):
		s.scanQuote(&tok)
	case isControl(s.char):
		s.scanControl(&tok)
	case isPattern(s.char):
		s.scanPattern(&tok)
	case isSelector(s.char):
		s.scanSelector(&tok)
	default:
		tok.Type = TokIllegal
	}
	switch tok.Type {
	case TokComma:
		s.skip(isBlank)
	case TokIllegal:
		if s.curr > pos {
			s.reset(pos)
			tok.Literal = s.scanIllegal(isControl)
		}
	default:
	}
	return tok
}

func (s *Scanner) reset(offset int) {
	c, z := utf8.DecodeRune(s.buffer[offset:])
	s.char = c
	s.curr = offset
	s.next = offset + z
}

func (s *Scanner) scanIllegal(isDelim func(r rune) bool) string {
	var buf bytes.Buffer
	for !s.isDone() && !isDelim(s.char) {
		buf.WriteRune(s.char)
		s.readRune()
	}
	return buf.String()
}

func (s *Scanner) scanUntil(accept func(r rune) bool) (string, bool) {
	var buf bytes.Buffer
	isDelim := func(r rune) bool {
		return isControl(r) || isOperator(r) || isBlank(r) || isSelector(r)
	}
	for !s.isDone() && !isDelim(s.char) {
		if !accept(s.char) {
			return buf.String(), false
		}
		buf.WriteRune(s.char)
		s.readRune()
	}
	return buf.String(), true
}

func (s *Scanner) scanSelector(tok *Token) {
	s.readRune()

	var buf bytes.Buffer
	for !s.isDone() && isLetter(s.char) {
		buf.WriteRune(s.char)
		s.readRune()
	}
	tok.Literal = buf.String()
	if kind, ok := selectors[tok.Literal]; ok {
		tok.Type = kind
	} else {
		tok.Type = TokIllegal
	}
}

func (s *Scanner) scanBase(tok *Token) {
	var (
		buf    bytes.Buffer
		accept func(r rune) bool
	)
	buf.WriteRune(s.char)
	s.readRune()
	switch s.char {
	case hex:
		accept = isHexa
	case oct:
		accept = isOctal
	case bin:
		accept = isBinary
	default:
		tok.Type = TokIllegal
		return
	}
	buf.WriteRune(s.char)
	s.readRune()
	for !s.isDone() {
		if s.char == underscore {
			ok := accept(s.prevRune()) && accept(s.nextRune())
			if !ok {
				tok.Literal = buf.String()
				tok.Type = TokIllegal
				return
			}
			s.readRune()
		}
		if !accept(s.char) {
			break
		}
		buf.WriteRune(s.char)
		s.readRune()
	}
	tok.Literal = buf.String()
	tok.Type = TokInteger
}

func (s *Scanner) scanNumber(tok *Token) {
	var buf bytes.Buffer
	if isSign(s.char) {
		buf.WriteRune(s.char)
		s.readRune()
	}
	if s.char == '0' && isDigit(s.nextRune()) {
		tok.Type = TokIllegal
		return
	}
	var kind rune
Loop:
	for !s.isDone() {
		switch {
		case s.char == minus:
			kind = scanDate(s, &buf)
			break Loop
		case s.char == colon:
			kind = scanTime(s, &buf)
			break Loop
		case s.char == underscore:
			ok := isDigit(s.prevRune()) && isDigit(s.nextRune())
			if !ok {
				tok.Literal = buf.String()
				tok.Type = TokIllegal
				return
			}
		case s.char == dot:
			kind = scanFraction(s, &buf)
			break Loop
		case s.char == 'e' || s.char == 'E':
			kind = scanExponent(s, &buf)
			break Loop
		case isDigit(s.char):
			buf.WriteRune(s.char)
		default:
			break Loop
		}
		s.readRune()
	}
	if kind == 0 {
		kind = TokInteger
	}
	tok.Type = kind
	tok.Literal = buf.String()
}

func scanDate(s *Scanner, buf *bytes.Buffer) rune {
	scan := func() bool {
		if s.char != minus {
			return false
		}
		buf.WriteRune(s.char)
		s.readRune()
		for i := 0; i < 2; i++ {
			if !isDigit(s.char) {
				return false
			}
			buf.WriteRune(s.char)
			s.readRune()
		}
		return true
	}
	if !scan() {
		return TokIllegal
	}
	if !scan() {
		return TokIllegal
	}
	if (s.char == space || s.char == 'T') && isDigit(s.nextRune()) {
		buf.WriteRune(s.char)
		s.readRune()
		if kind := scanTime(s, buf); kind == TokIllegal {
			return kind
		}
		if kind := scanTimezone(s, buf); kind == TokIllegal {
			return kind
		}
		return TokDateTime
	}
	return TokDate
}

func scanTime(s *Scanner, buf *bytes.Buffer) rune {
	scan := func(check bool) bool {
		if check && s.char != colon {
			return false
		}
		buf.WriteRune(s.char)
		s.readRune()
		for i := 0; i < 2; i++ {
			if !isDigit(s.char) {
				return false
			}
			buf.WriteRune(s.char)
			s.readRune()
		}
		return true
	}
	if s.char != colon {
		scan(false)
	}
	if !scan(true) {
		return TokIllegal
	}
	if !scan(true) {
		return TokIllegal
	}
	if s.char == dot {
		buf.WriteRune(s.char)
		s.readRune()
		n := buf.Len()
		for isDigit(s.char) {
			buf.WriteRune(s.char)
			s.readRune()
		}
		if diff := buf.Len() - n; diff > 9 {
			return TokIllegal
		}
	}
	return TokTime
}

func scanTimezone(s *Scanner, buf *bytes.Buffer) rune {
	if s.char == 'Z' {
		buf.WriteRune(s.char)
		s.readRune()
		return TokDateTime
	}
	if s.char != plus && s.char != minus {
		return TokDateTime
	}
	scan := func() bool {
		for i := 0; i < 2; i++ {
			if !isDigit(s.char) {
				return false
			}
			buf.WriteRune(s.char)
			s.readRune()
		}
		return true
	}
	buf.WriteRune(s.char)
	s.readRune()
	if !scan() {
		return TokIllegal
	}
	buf.WriteRune(s.char)
	if s.char != colon {
		return TokIllegal
	}
	s.readRune()
	if !scan() {
		return TokIllegal
	}
	return TokDateTime
}

func scanFraction(s *Scanner, buf *bytes.Buffer) rune {
	buf.WriteRune(s.char)
	s.readRune()
Loop:
	for !s.isDone() {
		switch {
		case s.char == 'e' || s.char == 'E':
			return scanExponent(s, buf)
		case s.char == underscore:
			ok := isDigit(s.prevRune()) && isDigit(s.nextRune())
			if !ok {
				return TokIllegal
			}
		case isDigit(s.char):
			buf.WriteRune(s.char)
		default:
			break Loop
		}
		s.readRune()
	}
	return TokFloat
}

func scanExponent(s *Scanner, buf *bytes.Buffer) rune {
	buf.WriteRune(s.char)
	s.readRune()
	if isSign(s.char) {
		buf.WriteRune(s.char)
		s.readRune()
	}
Loop:
	for !s.isDone() {
		switch {
		case s.char == underscore:
			ok := isDigit(s.prevRune()) && isDigit(s.nextRune())
			if !ok {
				return TokIllegal
			}
		case isDigit(s.char):
			buf.WriteRune(s.char)
		default:
			break Loop
		}
		s.readRune()
	}
	return TokFloat
}

func (s *Scanner) scanDigit(tok *Token) {
	str, ok := s.scanUntil(isDigit)
	tok.Literal = str
	tok.Type = TokInteger
	if !ok {
		tok.Type = TokIllegal
	}
}

func (s *Scanner) scanLiteral(tok *Token) {
	str, ok := s.scanUntil(isAlpha)
	tok.Literal = str
	tok.Type = TokLiteral
	if kind, ok := identifiers[tok.Literal]; ok {
		tok.Type = kind
	}
	if !ok {
		tok.Type = TokIllegal
	}
}

func (s *Scanner) scanPattern(tok *Token) {
	var buf bytes.Buffer
	s.readRune()
	for !s.isDone() && s.char != slash {
		buf.WriteRune(s.char)
		s.readRune()
	}
	tok.Literal = buf.String()
	tok.Type = TokPattern
	if s.char != slash {
		tok.Type = TokIllegal
		return
	}
	s.readRune()
}

func (s *Scanner) scanQuote(tok *Token) {
	var (
		buf   bytes.Buffer
		quote = s.char
	)
	s.readRune()
Loop:
	for !s.isDone() && s.char != quote {
		if quote == dquote && s.char == backslash {
			switch char := scanEscape(s); char {
			case utf8.RuneError:
				break Loop
			case 0:
				continue
			default:
				s.char = char
			}
		}
		buf.WriteRune(s.char)
		s.readRune()
	}
	tok.Literal = buf.String()
	tok.Type = TokLiteral
	if s.char != quote {
		tok.Type = TokIllegal
		return
	}
	s.readRune()
}

func scanEscape(s *Scanner) rune {
	if s.char == 'u' || s.char == 'U' {
		return scanUnicodeEscape(s)
	}
	if char, ok := escapes[s.char]; ok {
		s.readRune()
		return char
	}
	return utf8.RuneError
}

func scanUnicodeEscape(s *Scanner) rune {
	var (
		char   int32
		offset int32
		step   int32
	)
	if s.char == 'u' {
		step, offset = 4, 12
	} else {
		step, offset = 8, 28
	}
	for i := int32(0); i < step; i++ {
		s.readRune()
		var x rune
		switch {
		case s.char >= '0' && s.char <= '9':
			x = s.char - '0'
		case s.char >= 'a' && s.char <= 'f':
			x = s.char - 'a'
		case s.char >= 'A' && s.char <= 'F':
			x = s.char - 'A'
		default:
			return utf8.RuneError
		}
		char |= x << offset
		offset -= step
	}
	return char
}

func (s *Scanner) scanOperator(tok *Token) {
	var k rune
	switch s.char {
	case star:
		if s.nextRune() == equal {
			s.readRune()
			k = TokContains
		} else {
			k = TokIllegal
		}
	case tilde:
		if s.nextRune() == equal {
			s.readRune()
			k = TokMatch
		} else {
			k = TokIllegal
		}
	case dollar:
		if s.nextRune() == equal {
			s.readRune()
			k = TokEndsWith
		} else {
			k = TokIllegal
		}
	case caret:
		if s.nextRune() == equal {
			s.readRune()
			k = TokStartsWith
		} else {
			k = TokIllegal
		}
	case equal:
		if s.nextRune() == equal {
			s.readRune()
			k = TokEqual
		} else {
			k = TokIllegal
		}
	case bang:
		if s.nextRune() == equal {
			s.readRune()
			k = TokNotEqual
		} else {
			k = TokIllegal
		}
	case langle:
		k = TokLesser
		if s.nextRune() == equal {
			s.readRune()
			k = TokLessEq
		}
	case rangle:
		k = TokGreater
		if s.nextRune() == equal {
			s.readRune()
			k = TokGreatEq
		}
	case ampersand:
		s.readRune()
		if s.char == ampersand {
			k = TokAnd
		}
	case pipe:
		s.readRune()
		if s.char == pipe {
			k = TokOr
		}
	case comma:
		k = TokComma
	}
	s.readRune()
	tok.Type = k
}

func (s *Scanner) scanControl(tok *Token) {
	var k rune
	switch s.char {
	case lparen:
		k = TokBegGrp
	case rparen:
		k = TokEndGrp
	case comma:
		k = TokComma
	case lsquare:
		k = TokBegExpr
	case rsquare:
		k = TokEndExpr
	case percent:
		k = TokValue
	case dollar:
		k = TokRegular
	case arobase:
		k = TokArray
	case dot:
		k = TokLevelOne
		if s.nextRune() == dot {
			s.readRune()
			k = TokLevelAny
		}
		if k == TokLevelAny && s.nextRune() == bang {
			s.readRune()
			k = TokLevelGreedy
		}
	}
	tok.Type = k
	s.readRune()
}

func (s *Scanner) isDone() bool {
	return s.char == TokEOF || s.char == TokIllegal
}

func (s *Scanner) readRune() {
	if s.char == TokEOF {
		return
	}
	c, z := utf8.DecodeRune(s.buffer[s.next:])
	if c == utf8.RuneError {
		if z == 0 {
			s.char = TokEOF
		} else {
			s.char = TokIllegal
		}
		return
	}
	s.char, s.curr, s.next = c, s.next, s.next+z
}

func (s *Scanner) unreadRune() {
	s.next = s.curr
	s.curr = s.curr - utf8.RuneLen(s.char)
}

func (s *Scanner) nextRune() rune {
	c, _ := utf8.DecodeRune(s.buffer[s.next:])
	return c
}

func (s *Scanner) prevRune() rune {
	c, _ := utf8.DecodeLastRune(s.buffer[:s.curr])
	return c
}

func (s *Scanner) skip(fn func(rune) bool) {
	for fn(s.char) {
		s.readRune()
	}
}

func isSign(r rune) bool {
	return r == plus || r == minus
}

func isBlank(r rune) bool {
	return r == space || r == tab
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isAlpha(r rune) bool {
	return isLetter(r) || isDigit(r) || r == underscore || r == minus
}

func isQuote(r rune) bool {
	return r == squote || r == dquote
}

func isOperator(r rune) bool {
	return r == equal || r == bang || r == langle || r == rangle ||
		r == ampersand || r == pipe || r == tilde || r == caret ||
		r == dollar || r == star || r == comma
}

func isControl(r rune) bool {
	return r == percent || r == arobase || r == dollar || r == dot ||
		r == lsquare || r == rsquare || r == lparen || r == rparen || r == comma
}

func isPattern(r rune) bool {
	return r == slash
}

func isSelector(r rune) bool {
	return r == colon
}

func isBinary(r rune) bool {
	return r == '0' || r == '1'
}

func isOctal(r rune) bool {
	return r >= '0' && r <= '7'
}

func isHexa(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'Z')
}
