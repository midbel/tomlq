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
)

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
	switch {
	case isQuote(s.char):
		s.scanQuote(&tok)
	case isOperator(s.char):
		s.scanOperator(&tok)
	case isDigit(s.char) || s.char == minus || s.char == plus:
		s.scanNumber(&tok)
	case isAlpha(s.char):
		s.scanLiteral(&tok)
	case isControl(s.char):
		s.scanControl(&tok)
	case isPattern(s.char):
		s.scanPattern(&tok)
	default:
		tok.Type = TokIllegal
	}
	return tok
}

func (s *Scanner) scanDefault() Token {
	var tok Token
	if s.isDone() {
		tok.Type = s.char
		return tok
	}
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
	if tok.Type == TokComma {
		s.skip(isBlank)
	}
	return tok
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

func (s *Scanner) scanNumberIf(tok *Token, accept func(rune) bool) {
	var buf bytes.Buffer

	buf.WriteRune(s.char)
	s.readRune()
	buf.WriteRune(s.char)
	s.readRune()
	for !s.isDone() {
		if s.char == underscore {
			ok := accept(s.prevRune()) && accept(s.nextRune())
			if !ok {
				tok.Type = TokIllegal
				return
			}
			s.readRune()
			continue
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
	var (
		signed = s.char == plus || s.char == minus
		pos    = s.curr
	)
	if signed {
		s.readRune()
	}
	if s.char == '0' {
		switch peek := s.nextRune(); {
		case peek == 'x':
			s.scanNumberIf(tok, isHexa)
		case peek == 'o':
			s.scanNumberIf(tok, isOctal)
		case peek == 'b':
			s.scanNumberIf(tok, isBinary)
		case isDigit(peek):
			tok.Type = TokIllegal
		default:
		}
		if tok.Type == TokIllegal || tok.Type == TokInteger {
			if signed {
				tok.Type = TokIllegal
			}
			return
		}
		s.readRune()
	}
	var kind rune
	for kind == 0 && !isBlank(s.char) && !isOperator(s.char) && s.char != rparen && s.char != rsquare {
		switch s.char {
		case dot:
			kind = s.scanFraction()
		case 'e', 'E':
			kind = s.scanExponent()
		case colon:
			kind = s.scanTime()
		case minus:
			kind = s.scanDate()
		case underscore:
			ok := isDigit(s.prevRune()) && isDigit(s.nextRune())
			if !ok {
				kind = TokIllegal
			}
			s.readRune()
		default:
			s.readRune()
		}
	}
	if kind == 0 {
		kind = TokInteger
	}
	tok.Type = kind
	tok.Literal = string(s.buffer[pos:s.curr])
}

func (s *Scanner) scanFraction() rune {
	s.readRune()
	for !s.isDone() {
		switch {
		case s.char == underscore:
			ok := isDigit(s.prevRune()) && isDigit(s.nextRune())
			if !ok {
				return TokIllegal
			}
		case isDigit(s.char):
		case s.char == 'e' || s.char == 'E':
			return s.scanExponent()
		default:
			return TokFloat
		}
		s.readRune()
	}
	return TokFloat
}

func (s *Scanner) scanExponent() rune {
	s.readRune()
	switch {
	case s.char == minus || s.char == plus || isDigit(s.char):
		s.readRune()
	default:
		return TokIllegal
	}
	for isDigit(s.char) {
		if s.char == underscore {
			ok := isDigit(s.prevRune()) && isDigit(s.nextRune())
			if !ok {
				return TokIllegal
			}
		}
		s.readRune()
	}
	return TokFloat
}

func (s *Scanner) scanDate() rune {
	scan := func() bool {
		if s.char != minus {
			return false
		}
		s.readRune()
		for i := 0; i < 2; i++ {
			if !isDigit(s.char) {
				return false
			}
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
	if s.char == space || s.char == 'T' {
		s.readRune()
		if kind := s.scanTime(); kind == TokIllegal {
			return kind
		}
		if !s.scanTimezone() {
			return TokIllegal
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
		s.readRune()
		for i := 0; i < 2; i++ {
			if !isDigit(s.char) {
				return false
			}
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
		s.readRune()
		for isDigit(s.char) {
			s.readRune()
		}
	}
	return TokTime
}

func (s *Scanner) scanTimezone() bool {
	if s.char == 'Z' {
		s.readRune()
		return true
	}
	if s.char != plus && s.char != minus {
		return true
	}
	scan := func() bool {
		for i := 0; i < 2; i++ {
			if !isDigit(s.char) {
				return false
			}
			s.readRune()
		}
		return true
	}
	s.readRune()
	if !scan() {
		return false
	}
	if s.char != colon {
		return false
	}
	s.readRune()
	return scan()
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
	for !s.isDone() && s.char != quote {
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
		k = TokNot
		if s.nextRune() == equal {
			s.readRune()
			k = TokNotEqual
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
