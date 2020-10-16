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

type Scanner struct {
	input []byte
	char  rune
	curr  int
	next  int

	buf  bytes.Buffer
	scan func() Token
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
	case isDigit(s.char) || (isSign(s.char) && isDigit(s.nextRune())):
		k := s.nextRune()
		if s.char == zero && (k == hex || k == bin || k == oct) {
			s.scanBase(&tok)
		} else {
			s.scanNumber(&tok)
		}
	case isAlpha(s.char) || (isSign(s.char) && isLetter(s.nextRune())):
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
		s.scanIllegal(func(r rune) bool { return isControl(r) || isOperator(r) })
		tok.Literal = s.literal()
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
			s.scanIllegal(isControl)
			tok.Literal = s.literal()
		}
	default:
	}
	return tok
}

func (s *Scanner) scanIllegal(isDelim func(r rune) bool) {
	for !s.isDone() && !isDelim(s.char) {
		s.writeRune(s.char)
		s.readRune()
	}
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

func (s *Scanner) scanSelector(tok *Token) {
	s.readRune()
	for !s.isDone() && isLetter(s.char) {
		s.writeRune(s.char)
		s.readRune()
	}
	tok.Literal = s.literal()
	if kind, ok := selectors[tok.Literal]; ok {
		tok.Type = kind
	} else {
		tok.Type = TokIllegal
	}
}

func (s *Scanner) scanBase(tok *Token) {
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
		tok.Type = TokIllegal
		return
	}
	s.writeRune(s.char)
	s.readRune()
	for !s.isDone() {
		if s.char == underscore {
			ok := accept(s.prevRune()) && accept(s.nextRune())
			if !ok {
				tok.Literal = s.literal()
				tok.Type = TokIllegal
				return
			}
			s.readRune()
		}
		if !accept(s.char) {
			break
		}
		s.writeRune(s.char)
		s.readRune()
	}
	tok.Literal = s.literal()
	tok.Type = TokInteger
}

func (s *Scanner) scanNumber(tok *Token) {
	if isSign(s.char) {
		s.writeRune(s.char)
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
			kind = s.scanDate()
			break Loop
		case s.char == colon:
			kind = s.scanTime()
			break Loop
		case s.char == underscore:
			ok := isDigit(s.prevRune()) && isDigit(s.nextRune())
			if !ok {
				tok.Literal = s.literal()
				tok.Type = TokIllegal
				return
			}
		case s.char == dot:
			kind = s.scanFraction()
			break Loop
		case s.char == 'e' || s.char == 'E':
			kind = s.scanExponent()
			break Loop
		case isDigit(s.char):
			s.writeRune(s.char)
		default:
			break Loop
		}
		s.readRune()
	}
	if kind == 0 {
		kind = TokInteger
	}
	tok.Type = kind
	tok.Literal = s.literal()
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

func (s *Scanner) scanDigit(tok *Token) {
	ok := s.scanUntil(isDigit)
	tok.Literal = s.literal()
	tok.Type = TokInteger
	if !ok {
		tok.Type = TokIllegal
	}
}

func (s *Scanner) scanLiteral(tok *Token) {
	ok := s.scanUntil(isAlpha)
	tok.Literal = s.literal()
	tok.Type = TokLiteral
	if kind, ok := identifiers[tok.Literal]; ok {
		tok.Type = kind
	}
	if !ok {
		tok.Type = TokIllegal
	}
}

func (s *Scanner) scanPattern(tok *Token) {
	s.readRune()
	for !s.isDone() && s.char != slash {
		s.writeRune(s.char)
		s.readRune()
	}
	tok.Literal = s.literal()
	tok.Type = TokPattern
	if s.char != slash {
		tok.Type = TokIllegal
		return
	}
	s.readRune()
}

func (s *Scanner) scanQuote(tok *Token) {
	quote := s.char
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
		s.writeRune(s.char)
		s.readRune()
	}
	tok.Literal = s.literal()
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
