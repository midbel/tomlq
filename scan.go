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
	unicode4   = 'u'
	unicode8   = 'U'
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

type Scanner struct {
	input []byte
	char  rune
	curr  int
	next  int

	buf  bytes.Buffer
	scan func() rune
}

func NewScanner(str string) *Scanner {
	var s Scanner
	s.input = []byte(str)
	s.scan = s.scanDefault

	s.readRune()
	return &s
}

func (s *Scanner) Scan() Token {
	defer s.buf.Reset()
	kind := s.scan()
	switch kind {
	case TokBegExpr:
		s.scan = s.scanExpr
	case TokEndExpr:
		s.scan = s.scanDefault
	}
	return Token{
		Literal: s.literal(),
		Type:    kind,
	}
}

func (s *Scanner) scanExpr() rune {
	if s.isDone() {
		return TokIllegal
	}
	s.skip(isBlank)
	var (
		pos = s.curr
		tok rune
	)
	switch {
	case isQuote(s.char):
		tok = s.scanQuote()
	case isOperator(s.char):
		tok = s.scanOperator()
	case isDigit(s.char) || (isSign(s.char) && isDigit(s.nextRune())):
		k := s.nextRune()
		if s.char == zero && (k == hex || k == bin || k == oct) {
			tok = s.scanBase()
		} else {
			tok = s.scanNumber()
		}
	case isAlpha(s.char) || (isSign(s.char) && isLetter(s.nextRune())):
		tok = s.scanLiteral()
	case isControl(s.char):
		tok = s.scanControl()
	case isPattern(s.char):
		tok = s.scanPattern()
	default:
		tok = TokIllegal
	}
	if tok == TokIllegal && s.curr > pos {
		s.reset(pos)
		s.scanIllegal(func(r rune) bool { return isControl(r) || isOperator(r) })
	}
	return tok
}

func (s *Scanner) scanDefault() rune {
	if s.isDone() {
		return TokEOF
	}
	var (
		pos = s.curr
		tok rune
	)
	switch {
	case isDigit(s.char):
		tok = s.scanDigit()
	case isLetter(s.char):
		tok = s.scanLiteral()
	case isQuote(s.char):
		tok = s.scanQuote()
	case isControl(s.char):
		tok = s.scanControl()
	case isPattern(s.char):
		tok = s.scanPattern()
	case isSelector(s.char):
		tok = s.scanSelector()
	default:
		tok = TokIllegal
	}
	switch tok {
	case TokComma:
		s.skip(isBlank)
	case TokIllegal:
		if s.curr > pos {
			s.reset(pos)
			s.scanIllegal(isControl)
		}
	default:
	}
	return tok
}

func (s *Scanner) scanIllegal(isDelim func(r rune) bool) rune {
	for !s.isDone() && !isDelim(s.char) {
		s.writeRune(s.char)
		s.readRune()
	}
	return TokIllegal
}

func (s *Scanner) scanUntil(accept func(r rune) bool) bool {
	isDelim := func(r rune) bool {
		return isControl(r) || isOperator(r) || isBlank(r) || isSelector(r)
	}
	for !s.isDone() && !isDelim(s.char) {
		if !accept(s.char) {
			return false
		}
		s.writeRune(s.char)
		s.readRune()
	}
	return true
}

func (s *Scanner) scanSelector() rune {
	s.readRune()
	for !s.isDone() && isLetter(s.char) {
		s.writeRune(s.char)
		s.readRune()
	}
	if kind, ok := selectors[s.literal()]; ok {
		return kind
	}
	return TokIllegal
}

func (s *Scanner) scanBase() rune {
	var accept func(r rune) bool
	s.writeRune(s.char)
	s.readRune()
	switch s.char {
	case hex:
		accept = isHexa
	case oct:
		accept = isOctal
	case bin:
		accept = isBinary
	default:
		return TokIllegal
	}
	s.writeRune(s.char)
	s.readRune()
	for !s.isDone() {
		if s.char == underscore {
			ok := accept(s.prevRune()) && accept(s.nextRune())
			if !ok {
				return TokIllegal
			}
			s.readRune()
		}
		if !accept(s.char) {
			break
		}
		s.writeRune(s.char)
		s.readRune()
	}
	return TokInteger
}

func (s *Scanner) scanNumber() rune {
	if isSign(s.char) {
		s.writeRune(s.char)
		s.readRune()
	}
	if s.char == '0' && isDigit(s.nextRune()) {
		return TokIllegal
	}
Loop:
	for !s.isDone() {
		switch {
		case s.char == minus:
			return s.scanDate()
		case s.char == colon:
			return s.scanTime()
		case s.char == underscore:
			ok := isDigit(s.prevRune()) && isDigit(s.nextRune())
			if !ok {
				return TokIllegal
			}
		case s.char == dot:
			return s.scanFraction()
		case s.char == 'e' || s.char == 'E':
			return s.scanExponent()
		case isDigit(s.char):
			s.writeRune(s.char)
		default:
			break Loop
		}
		s.readRune()
	}
	return TokInteger
}

func (s *Scanner) scanDate() rune {
	scan := func() bool {
		if s.char != minus {
			return false
		}
		s.writeRune(s.char)
		s.readRune()
		for i := 0; i < 2; i++ {
			if !isDigit(s.char) {
				return false
			}
			s.writeRune(s.char)
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
		s.writeRune(s.char)
		s.readRune()
		if kind := s.scanTime(); kind == TokIllegal {
			return kind
		}
		if kind := s.scanTimezone(); kind == TokIllegal {
			return kind
		}
		return TokDateTime
	}
	return TokDate
}

func (s *Scanner) scanTime() rune {
	scan := func(check bool) bool {
		if check && s.char != colon {
			return false
		}
		s.writeRune(s.char)
		s.readRune()
		for i := 0; i < 2; i++ {
			if !isDigit(s.char) {
				return false
			}
			s.writeRune(s.char)
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
		s.writeRune(s.char)
		s.readRune()
		n := s.written()
		for isDigit(s.char) {
			s.writeRune(s.char)
			s.readRune()
		}
		if diff := s.written() - n; diff > 9 {
			return TokIllegal
		}
	}
	return TokTime
}

func (s *Scanner) scanTimezone() rune {
	if s.char == 'Z' {
		s.writeRune(s.char)
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
			s.writeRune(s.char)
			s.readRune()
		}
		return true
	}
	s.writeRune(s.char)
	s.readRune()
	if !scan() {
		return TokIllegal
	}
	s.writeRune(s.char)
	if s.char != colon {
		return TokIllegal
	}
	s.readRune()
	if !scan() {
		return TokIllegal
	}
	return TokDateTime
}

func (s *Scanner) scanFraction() rune {
	s.writeRune(s.char)
	s.readRune()
Loop:
	for !s.isDone() {
		switch {
		case s.char == 'e' || s.char == 'E':
			return s.scanExponent()
		case s.char == underscore:
			ok := isDigit(s.prevRune()) && isDigit(s.nextRune())
			if !ok {
				return TokIllegal
			}
		case isDigit(s.char):
			s.writeRune(s.char)
		default:
			break Loop
		}
		s.readRune()
	}
	return TokFloat
}

func (s *Scanner) scanExponent() rune {
	s.writeRune(s.char)
	s.readRune()
	if isSign(s.char) {
		s.writeRune(s.char)
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
			s.writeRune(s.char)
		default:
			break Loop
		}
		s.readRune()
	}
	return TokFloat
}

func (s *Scanner) scanDigit() rune {
	ok := s.scanUntil(isDigit)
	if !ok {
		return TokIllegal
	}
	return TokInteger
}

func (s *Scanner) scanLiteral() rune {
	ok := s.scanUntil(isAlpha)
	if kind, ok := identifiers[s.literal()]; ok {
		return kind
	}
	if !ok {
		return TokIllegal
	}
	return TokLiteral
}

func (s *Scanner) scanPattern() rune {
	s.readRune()
	for !s.isDone() && s.char != slash {
		s.writeRune(s.char)
		s.readRune()
	}
	if s.char != slash {
		return TokIllegal
	}
	s.readRune()
	return TokPattern
}

func (s *Scanner) scanQuote() rune {
	quote := s.char
	s.readRune()
	for !s.isDone() && s.char != quote {
		if quote == dquote && s.char == backslash {
			switch char := s.scanEscape(); char {
			case utf8.RuneError:
				return TokIllegal
			case 0:
				continue
			default:
				s.char = char
			}
		}
		s.writeRune(s.char)
		s.readRune()
	}
	if s.char != quote {
		return TokIllegal
	}
	s.readRune()
	return TokLiteral
}

func (s *Scanner) scanEscape() rune {
	if s.char == unicode4 || s.char == unicode8 {
		return s.scanUnicodeEscape()
	}
	if char, ok := escapes[s.char]; ok {
		s.readRune()
		return char
	}
	return utf8.RuneError
}

func (s *Scanner) scanUnicodeEscape() rune {
	var (
		char   int32
		offset int32
		step   int32
	)
	if s.char == unicode4 {
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

func (s *Scanner) scanOperator() rune {
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
	return k
}

func (s *Scanner) scanControl() rune {
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
	s.readRune()
	return k
}

func (s *Scanner) isDone() bool {
	return s.char == TokEOF || s.char == TokIllegal
}

func (s *Scanner) reset(at int) {
	c, z := utf8.DecodeRune(s.input[at:])
	s.char = c
	s.curr = at
	s.next = at + z
}

func (s *Scanner) literal() string {
	return s.buf.String()
}

func (s *Scanner) written() int {
	return s.buf.Len()
}

func (s *Scanner) writeRune(char rune) {
	s.buf.WriteRune(char)
}

func (s *Scanner) readRune() {
	if s.char == TokEOF {
		return
	}
	c, z := utf8.DecodeRune(s.input[s.next:])
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
	c, _ := utf8.DecodeRune(s.input[s.next:])
	return c
}

func (s *Scanner) prevRune() rune {
	c, _ := utf8.DecodeLastRune(s.input[:s.curr])
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
