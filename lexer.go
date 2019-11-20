package sexpr

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"text/scanner"
)

var (
	wordStart = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_")
	wordBody  = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_")

	wordBreak = []rune{'(', ')', '"', '#', ' ', '\r', '\t', '\f', '\n', '[', ']'}
)

type token struct {
	tt  tokenType
	val string

	col  uint64
	line uint64
	pos  uint64
}

func (t token) String() string {
	return fmt.Sprintf("%q", t.val)
	return fmt.Sprintf("%q %d:%d (%s)", t.val, t.col, t.line, tokenName(t.tt))
}

type lexState func(*lexer) lexState

type tokenType uint8

const (
	tokenInvalid tokenType = iota

	tokenOpenExpression
	tokenCloseExpression
	tokenNewLine
	tokenQuote
	tokenHash
	tokenSeparator
	tokenAtom
	tokenInteger
	tokenFloat
	tokenString
	tokenBool
	tokenNil
	tokenColon
	tokenDot
	tokenBackslash
	tokenOpenMap
	tokenCloseMap
	tokenOpenList
	tokenCloseList

	tokenEOF
)

var tokenValues = map[tokenType][]rune{
	tokenOpenMap:         []rune{'{'},
	tokenCloseMap:        []rune{'}'},
	tokenOpenExpression:  []rune{'('},
	tokenCloseExpression: []rune{')'},
	tokenOpenList:        []rune{'['},
	tokenCloseList:       []rune{']'},
	tokenQuote:           []rune{'"'},
	tokenDot:             []rune{'.'},
	tokenColon:           []rune{':'},
	tokenBackslash:       []rune{'\\'},
	tokenHash:            []rune{'#'},
	tokenNewLine:         []rune{'\n'},
	tokenSeparator:       []rune{' ', '\r', '\t', '\f', '\n'},
}

var tokenNames = map[tokenType]string{
	tokenOpenExpression:  "[open expr]",
	tokenCloseExpression: "[close expr]",
	tokenOpenList:        "[open list]",
	tokenCloseList:       "[close list]",
	tokenOpenMap:         "[open map]",
	tokenCloseMap:        "[close map]",
	tokenQuote:           "[quote]",
	tokenHash:            "[hash]",
	tokenNewLine:         "[newline]",
	tokenAtom:            "[atom]",
	tokenDot:             "[dot]",
	tokenString:          "[string]",
	tokenBool:            "[bool]",
	tokenInteger:         "[integer]",
	tokenFloat:           "[float]",
	tokenSeparator:       "[separator]",

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
	isOpenExpression  = isTokenType(tokenOpenExpression)
	isCloseExpression = isTokenType(tokenCloseExpression)
	isOpenList        = isTokenType(tokenOpenList)
	isCloseList       = isTokenType(tokenCloseList)
	isOpenMap         = isTokenType(tokenOpenMap)
	isCloseMap        = isTokenType(tokenCloseMap)
	isSeparator       = isTokenType(tokenSeparator)
	isQuote           = isTokenType(tokenQuote)
	isHash            = isTokenType(tokenHash)
	isDot             = isTokenType(tokenDot)
	isNewLine         = isTokenType(tokenNewLine)
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
		done:   make(chan struct{}),
		buf:    []rune{},
	}
}

type lexer struct {
	in *scanner.Scanner

	tokens chan token
	done   chan struct{}

	buf  []rune
	col  uint64
	line uint64
	pos  uint64
}

func (lx *lexer) stop() {
	for {
		select {
		case <-lx.tokens:
			// drain channel
		default:
			lx.done <- struct{}{}
			close(lx.tokens)
			return
		}
	}
}

func (lx *lexer) run() error {

	for state := lexDefaultState; state != nil; {
		select {
		case <-lx.done:
			return nil
		default:
			state = state(lx)
		}
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
	if err != nil {
		return lexStateError(err)
	}

	switch {
	case isOpenExpression(r):
		return lexOpenExpressionState
	case isCloseExpression(r):
		return lexCloseExpressionState
	case isOpenList(r):
		return lexOpenListState
	case isCloseList(r):
		return lexCloseListState
	case isQuote(r):
		return lexQuote
	case isHash(r):
		return lexHash
	case isNewLine(r):
		return lexNewLine
	case isSeparator(r):
		return lexSeparator
	default:
		return lexAtom
	}

	panic("unreachable")
}

func lexQuote(lx *lexer) lexState {
	lx.emit(tokenQuote)
	return lexDefaultState
}

func lexNewLine(lx *lexer) lexState {
	lx.emit(tokenNewLine)
	return lexDefaultState
}

func lexHash(lx *lexer) lexState {
	lx.emit(tokenHash)
	return lexDefaultState
}

func lexOpenListState(lx *lexer) lexState {
	lx.emit(tokenOpenList)
	return lexDefaultState
}

func lexCloseListState(lx *lexer) lexState {
	lx.emit(tokenCloseList)
	return lexDefaultState
}

func lexOpenExpressionState(lx *lexer) lexState {
	lx.emit(tokenOpenExpression)
	return lexDefaultState
}

func lexCloseExpressionState(lx *lexer) lexState {
	lx.emit(tokenCloseExpression)
	return lexDefaultState
}

func lexSeparator(lx *lexer) lexState {
	for isSeparator(lx.peek()) {
		if _, err := lx.next(); err != nil {
			return lexStateError(err)
		}
	}
	lx.emit(tokenSeparator)
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
		log.Printf("lexer error: %v", err)
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
	if err != nil {
		return nil, err
	}

	return tokens, nil
}
