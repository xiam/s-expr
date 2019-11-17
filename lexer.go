package sexpr

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"text/scanner"
)

var (
	wordBreak = []rune{'(', ')', '"', '#', ' ', '\r', '\t', '\f', '\n'}
)

type token struct {
	tt  tokenType
	val string

	col  uint64
	line uint64
	pos  uint64
}

func (t token) String() string {
	return fmt.Sprintf("%q %d:%d (%s)", t.val, t.col, t.line, tokenName(t.tt))
}

type lexState func(*lexer) lexState

type tokenType uint8

const (
	tokenOpenList tokenType = iota
	tokenCloseList
	tokenNewLine
	tokenQuote
	tokenHash
	tokenAtomSeparator
	tokenAtom

	tokenInvalid
	tokenEOF
)

var tokenValues = map[tokenType][]rune{
	tokenOpenList:      []rune{'('},
	tokenCloseList:     []rune{')'},
	tokenQuote:         []rune{'"'},
	tokenHash:          []rune{'#'},
	tokenNewLine:       []rune{'\n'},
	tokenAtomSeparator: []rune{' ', '\r', '\t', '\f', '\n'},
}

var tokenNames = map[tokenType]string{
	tokenOpenList:      "[open parenthesis]",
	tokenCloseList:     "[close parenthesis]",
	tokenQuote:         "[quote]",
	tokenHash:          "[hash]",
	tokenNewLine:       "[newline]",
	tokenAtom:          "[atom]",
	tokenAtomSeparator: "[atom separator]",

	tokenEOF:     "[EOF]",
	tokenInvalid: "[invalid]",
}

func tokenName(tt tokenType) string {
	if v, ok := tokenNames[tt]; ok {
		return v
	}
	return tokenNames[tokenInvalid]
}

func isTokenType(tt tokenType) func(r rune) bool {
	return func(r rune) bool {
		for _, v := range tokenValues[tt] {
			if v == r {
				return true
			}
		}
		return false
	}
}

var (
	isOpenList      = isTokenType(tokenOpenList)
	isCloseList     = isTokenType(tokenCloseList)
	isAtomSeparator = isTokenType(tokenAtomSeparator)
	isQuote         = isTokenType(tokenQuote)
	isHash          = isTokenType(tokenHash)
	isNewLine       = isTokenType(tokenNewLine)
)

func isWordBreak(r rune) bool {
	for _, v := range wordBreak {
		if v == r {
			return true
		}
	}
	return false
}

func newLexer(r io.Reader) *lexer {
	s := &scanner.Scanner{
		Mode: scanner.ScanIdents | scanner.ScanFloats | scanner.ScanChars | scanner.ScanStrings | scanner.ScanRawStrings | scanner.ScanComments,
	}

	return &lexer{
		in:     s.Init(r),
		tokens: make(chan token),
		buf:    []rune{},
	}
}

type lexer struct {
	in     *scanner.Scanner
	tokens chan token

	buf  []rune
	col  uint64
	line uint64
	pos  uint64
}

func (lx *lexer) run() error {
	for state := lexDefaultState; state != nil; {
		state = state(lx)
	}

	lx.emit(tokenEOF)
	close(lx.tokens)

	return nil
}

func (lx *lexer) emit(tt tokenType) {
	lx.tokens <- token{
		val:  string(lx.buf),
		tt:   tt,
		col:  lx.col,
		line: lx.line,
		pos:  lx.pos,
	}
	lx.buf = []rune{}
}

func (lx *lexer) peek() rune {
	return lx.in.Peek()
}

func (lx *lexer) next() (rune, error) {
	r := lx.in.Next()
	if r == scanner.EOF {
		return rune(0), io.EOF
	}
	lx.buf = append(lx.buf, r)

	lx.pos++
	if lx.line == 0 || isNewLine(r) {
		lx.col = 0
		lx.line++
	}
	lx.col++
	return r, nil
}

func lexDefaultState(lx *lexer) lexState {
	r, err := lx.next()
	log.Printf("r: %s, err: %v", string(r), err)

	if err != nil {
		return lexStateError(err)
	}

	switch {
	case isOpenList(r):
		return lexOpenListState
	case isCloseList(r):
		return lexCloseListState
	case isAtomSeparator(r):
		return lexAtomSeparator
	default:
		return lexAtom
	}

	panic("unreachable")
}

func lexOpenListState(lx *lexer) lexState {
	lx.emit(tokenOpenList)
	return lexDefaultState
}

func lexCloseListState(lx *lexer) lexState {
	lx.emit(tokenCloseList)
	return lexDefaultState
}

func lexAtomSeparator(lx *lexer) lexState {
	for isAtomSeparator(lx.peek()) {
		if _, err := lx.next(); err != nil {
			return lexStateError(err)
		}
	}
	lx.emit(tokenAtomSeparator)
	return lexDefaultState
}

func lexAtom(lx *lexer) lexState {
	for !isWordBreak(lx.peek()) {
		if _, err := lx.next(); err != nil {
			return lexStateError(err)
		}
	}
	lx.emit(tokenAtom)
	return lexDefaultState
}

func lexStateError(err error) lexState {
	if err == io.EOF {
		return nil
	}
	return func(lx *lexer) lexState {
		return nil
	}
}

func lexStateEOF(lx *lexer) lexState {
	return nil
}

func tokenize(in []byte) ([]token, error) {
	tokens := []token{}
	errCh := make(chan error)

	lx := newLexer(bytes.NewReader(in))

	go func() {
		errCh <- lx.run()
	}()

	for tok := range lx.tokens {
		tokens = append(tokens, tok)
	}
	err := <-errCh
	return tokens, err
}
