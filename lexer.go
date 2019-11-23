package sexpr

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"text/scanner"
)

type token struct {
	tt   tokenType
	text string

	col  int
	line int
}

func (t token) String() string {
	return fmt.Sprintf("%q", t.text)
	return fmt.Sprintf("%q %d:%d (%s)", t.text, t.col, t.line, tokenName(t.tt))
}

type lexState func(*lexer) lexState

type tokenType uint8

const (
	tokenInvalid tokenType = iota

	tokenOpenList
	tokenCloseList

	tokenOpenMap
	tokenCloseMap

	tokenOpenExpression
	tokenCloseExpression
	tokenExpression

	tokenNewLine
	tokenQuote
	tokenHash
	tokenWhitespace

	tokenWord
	tokenInteger
	tokenSymbol
	tokenString

	tokenColon
	tokenStar
	tokenPercent
	tokenDot

	tokenBackslash

	tokenEOF
)

var tokenValues = map[tokenType][]rune{
	tokenOpenList:  []rune{'['},
	tokenCloseList: []rune{']'},

	tokenOpenMap:  []rune{'{'},
	tokenCloseMap: []rune{'}'},

	tokenOpenExpression:  []rune{'('},
	tokenCloseExpression: []rune{')'},

	tokenNewLine:    []rune{'\n'},
	tokenQuote:      []rune{'"'},
	tokenHash:       []rune{'#'},
	tokenWhitespace: []rune(" \f\t\r"),

	tokenWord:    []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ@_"),
	tokenInteger: []rune("0123456789"),

	tokenColon:   []rune{':'},
	tokenStar:    []rune{'*'},
	tokenPercent: []rune{'%'},
	tokenDot:     []rune{'.'},

	tokenBackslash: []rune{'\\'},
}

var tokenNames = map[tokenType]string{
	tokenInvalid: "[invalid]",

	tokenOpenList:  "[open list]",
	tokenCloseList: "[close list]",

	tokenOpenMap:  "[open map]",
	tokenCloseMap: "[close map]",

	tokenOpenExpression:  "[open expr]",
	tokenCloseExpression: "[close expr]",
	tokenExpression:      "[expression]",

	tokenNewLine:    "[newline]",
	tokenQuote:      "[quote]",
	tokenHash:       "[hash]",
	tokenWhitespace: "[separator]",

	tokenWord:    "[word]",
	tokenInteger: "[integer]",

	tokenColon:   "[colon]",
	tokenStar:    "[star]",
	tokenPercent: "[percent]",
	tokenDot:     "[dot]",

	tokenBackslash: "[backslash]",

	tokenEOF: "[EOF]",
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
	isOpenList  = isTokenType(tokenOpenList)
	isCloseList = isTokenType(tokenCloseList)

	isOpenMap  = isTokenType(tokenOpenMap)
	isCloseMap = isTokenType(tokenCloseMap)

	isOpenExpression  = isTokenType(tokenOpenExpression)
	isCloseExpression = isTokenType(tokenCloseExpression)

	isNewLine    = isTokenType(tokenNewLine)
	isQuote      = isTokenType(tokenQuote)
	isHash       = isTokenType(tokenHash)
	isWhitespace = isTokenType(tokenWhitespace)

	isWord    = isTokenType(tokenWord)
	isInteger = isTokenType(tokenInteger)

	isColon   = isTokenType(tokenColon)
	isStar    = isTokenType(tokenStar)
	isPercent = isTokenType(tokenPercent)
	isDot     = isTokenType(tokenDot)

	isBackslash = isTokenType(tokenBackslash)
)

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

	done chan struct{}

	buf []rune

	start  int
	offset int
	lines  int
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
		tt:   tt,
		text: string(lx.buf),

		col:  lx.start + 1,
		line: lx.lines + 1,
	}

	lx.start = lx.offset
	lx.buf = lx.buf[0:0]

	if tt == tokenNewLine {
		lx.lines++
		lx.start = 0
		lx.offset = 0
	}
}

func (lx *lexer) peek() rune {
	return lx.in.Peek()
}

func (lx *lexer) next() (rune, error) {
	lx.offset++

	r := lx.in.Next()
	if r == scanner.EOF {
		return rune(0), io.EOF
	}

	lx.buf = append(lx.buf, r)
	return r, nil
}

func lexDefaultState(lx *lexer) lexState {
	r, err := lx.next()
	if err != nil {
		return lexStateError(err)
	}

	switch {

	case isOpenList(r):
		return lexEmit(tokenOpenList)
	case isCloseList(r):
		return lexEmit(tokenCloseList)

	case isOpenMap(r):
		return lexEmit(tokenOpenMap)
	case isCloseMap(r):
		return lexEmit(tokenCloseMap)

	case isOpenExpression(r):
		return lexEmit(tokenOpenExpression)
	case isCloseExpression(r):
		return lexEmit(tokenCloseExpression)

	case isQuote(r):
		return lexEmit(tokenQuote)
	case isHash(r):
		return lexEmit(tokenHash)
	case isNewLine(r):
		return lexEmit(tokenNewLine)
	case isWhitespace(r):
		return lexCollectStream(tokenWhitespace)

	case isWord(r):
		return lexCollectStream(tokenWord)
	case isInteger(r):
		return lexCollectStream(tokenInteger)

	case isColon(r):
		return lexEmit(tokenColon)
	case isStar(r):
		return lexEmit(tokenStar)
	case isPercent(r):
		return lexEmit(tokenPercent)
	case isDot(r):
		return lexEmit(tokenDot)
	case isBackslash(r):
		return lexEmit(tokenBackslash)

	default:
		return lexString

	}

	panic("unreachable")
}

func lexString(lx *lexer) lexState {
	for {
		p := lx.peek()
		if isWhitespace(p) || isNewLine(p) {
			break
		}
		if _, err := lx.next(); err != nil {
			return lexStateError(err)
		}
	}
	lx.emit(tokenString)
	return lexDefaultState
}

func lexEmit(tt tokenType) lexState {
	return func(lx *lexer) lexState {
		lx.emit(tt)
		return lexDefaultState
	}
}

func lexCollectStream(tt tokenType) lexState {
	return func(lx *lexer) lexState {
		for (isTokenType(tt))(lx.peek()) {
			if _, err := lx.next(); err != nil {
				return lexStateError(err)
			}
		}
		return lexEmit(tt)
	}
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
