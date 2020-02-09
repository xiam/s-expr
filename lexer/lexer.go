package lexer

import (
	"bytes"
	"fmt"
	"io"
	"text/scanner"
)

var (
	ticket = struct{}{}
)

type lexState func(*Lexer) lexState

var (
	isOpenList  = isTokenType(TokenOpenList)
	isCloseList = isTokenType(TokenCloseList)

	isOpenMap  = isTokenType(TokenOpenMap)
	isCloseMap = isTokenType(TokenCloseMap)

	isOpenExpression  = isTokenType(TokenOpenExpression)
	isCloseExpression = isTokenType(TokenCloseExpression)

	isNewLine     = isTokenType(TokenNewLine)
	isDoubleQuote = isTokenType(TokenDoubleQuote)
	isHash        = isTokenType(TokenHash)
	isWhitespace  = isTokenType(TokenWhitespace)

	isWord    = isTokenType(TokenWord)
	isInteger = isTokenType(TokenInteger)

	isColon     = isTokenType(TokenColon)
	isDot       = isTokenType(TokenDot)
	isBackslash = isTokenType(TokenBackslash)
)

// New initializes a Lexer object
func New(r io.Reader) *Lexer {
	s := &scanner.Scanner{
		Mode: scanner.ScanIdents | scanner.ScanFloats | scanner.ScanChars | scanner.ScanStrings | scanner.ScanRawStrings | scanner.ScanComments,
	}

	return &Lexer{
		in:       s.Init(r),
		tickets:  make(chan struct{}),
		scanning: make(chan struct{}),
		tokens:   make(chan *Token),
		buf:      []rune{},
	}
}

// Lexer represents a lexical analyzer
type Lexer struct {
	in *scanner.Scanner

	lastTok *Token
	tokens  chan *Token

	tickets chan struct{}

	scanning chan struct{}

	lastErr error
	closed  bool

	buf []rune

	start  int
	offset int
	lines  int
}

// Next sends a signal to the Scan method for it to continue scanning
func (lx *Lexer) Next() bool {
	if lx.closed {
		return false
	}

	lx.tickets <- struct{}{}

	tok, _ := <-lx.tokens
	lx.lastTok = tok

	if tok.tt == TokenEOF {
		lx.closed = true
	}

	return true
}

// Token returns the most recent scanned token
func (lx *Lexer) Token() *Token {
	return lx.lastTok
}

// Stop requests the Scan method to stop scanning
func (lx *Lexer) Stop() {
	lx.closed = true
	close(lx.tickets)
	close(lx.scanning)
}

// Scan starts scanning the reader for tokens.
func (lx *Lexer) Scan() error {
	for state := lexDefaultState; state != nil; {
		select {
		case <-lx.scanning:
			return ErrForceStopped
		default:
			state = state(lx)
		}
	}

	if lx.lastErr == nil {
		lx.emit(TokenEOF)
	}

	return lx.lastErr
}

func (lx *Lexer) emit(tt TokenType) {
	_, ok := <-lx.tickets
	if !ok {
		return
	}

	tok := Token{
		tt:     tt,
		lexeme: string(lx.buf),

		col:  lx.start + 1,
		line: lx.lines + 1,
	}

	lx.start = lx.offset
	lx.buf = lx.buf[0:0]

	if tt == TokenNewLine {
		lx.lines++
		lx.start = 0
		lx.offset = 0
	}

	lx.tokens <- &tok
}

func (lx *Lexer) peek() rune {
	return lx.in.Peek()
}

func (lx *Lexer) next() (rune, error) {
	lx.offset++

	r := lx.in.Next()
	if r == scanner.EOF {
		return rune(0), io.EOF
	}

	lx.buf = append(lx.buf, r)
	return r, nil
}

func lexDefaultState(lx *Lexer) lexState {
	r, err := lx.next()
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return lexStateError(err)
	}

	switch {
	case isOpenList(r):
		return lexEmit(TokenOpenList)
	case isCloseList(r):
		return lexEmit(TokenCloseList)

	case isOpenMap(r):
		return lexEmit(TokenOpenMap)
	case isCloseMap(r):
		return lexEmit(TokenCloseMap)

	case isOpenExpression(r):
		return lexEmit(TokenOpenExpression)
	case isCloseExpression(r):
		return lexEmit(TokenCloseExpression)

	case isDoubleQuote(r):
		return lexEmit(TokenDoubleQuote)
	case isHash(r):
		return lexEmit(TokenHash)
	case isNewLine(r):
		return lexEmit(TokenNewLine)
	case isWhitespace(r):
		return lexCollectStream(TokenWhitespace)

	case isWord(r):
		return lexCollectStream(TokenWord)

	case isInteger(r):
		return lexCollectStream(TokenInteger)

	case isColon(r):
		return lexEmit(TokenColon)
	case isDot(r):
		return lexEmit(TokenDot)
	case isBackslash(r):
		return lexEmit(TokenBackslash)

	default:
		if isAritmeticSign(r) {
			return lexNumeric
		}
		return lexSequence
	}

	panic("unreachable")
}

func lexNumeric(lx *Lexer) lexState {
	p := lx.peek()
	if !isInteger(p) {
		return lexSequence
	}
	if _, err := lx.next(); err != nil {
		if err != io.EOF {
			return lexStateError(err)
		}
	}

	lx.emit(TokenInteger)
	return lexDefaultState
}

func lexSequence(lx *Lexer) lexState {
loop:
	for {
		p := lx.peek()
		switch {
		case isWhitespace(p), isNewLine(p), isDoubleQuote(p), isOpenList(p), isCloseList(p), isOpenExpression(p), isCloseExpression(p), isOpenMap(p), isCloseMap(p):
			break loop
		}
		if _, err := lx.next(); err != nil {
			if err == io.EOF {
				break
			}
			return lexStateError(err)
		}
	}
	lx.emit(TokenSequence)
	return lexDefaultState
}

func lexEmit(tt TokenType) lexState {
	return func(lx *Lexer) lexState {
		lx.emit(tt)
		return lexDefaultState
	}
}

func lexCollectStream(tt TokenType) lexState {
	return func(lx *Lexer) lexState {
		for (isTokenType(tt))(lx.peek()) {
			if _, err := lx.next(); err != nil {
				if err == io.EOF {
					break
				}
				return lexStateError(err)
			}
		}
		return lexEmit(tt)
	}
}

func lexStateError(err error) lexState {
	return func(lx *Lexer) lexState {
		lx.lastErr = fmt.Errorf("read error: %v", err)
		return nil
	}
}

// Tokenize takes an array of bytes and returns all the tokens within it,
// or an error if a token can't be identified.
func Tokenize(in []byte) ([]Token, error) {
	tokens := []Token{}
	done := make(chan struct{})

	lx := New(bytes.NewReader(in))

	go func() {
		for lx.Next() {
			tok := lx.Token()
			if tok == nil {
				break
			}
			tokens = append(tokens, *tok)
		}
		done <- struct{}{}
	}()

	if err := lx.Scan(); err != nil {
		return nil, err
	}

	<-done
	return tokens, nil
}
