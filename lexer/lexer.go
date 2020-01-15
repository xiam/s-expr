package lexer

import (
	"bytes"
	"io"
	"log"
	"text/scanner"
)

type lexState func(*Lexer) lexState

var (
	isOpenList  = isTokenType(TokenOpenList)
	isCloseList = isTokenType(TokenCloseList)

	isOpenMap  = isTokenType(TokenOpenMap)
	isCloseMap = isTokenType(TokenCloseMap)

	isOpenExpression  = isTokenType(TokenOpenExpression)
	isCloseExpression = isTokenType(TokenCloseExpression)

	isNewLine    = isTokenType(TokenNewLine)
	isQuote      = isTokenType(TokenQuote)
	isHash       = isTokenType(TokenHash)
	isWhitespace = isTokenType(TokenWhitespace)

	isWord    = isTokenType(TokenWord)
	isInteger = isTokenType(TokenInteger)

	isColon   = isTokenType(TokenColon)
	isStar    = isTokenType(TokenStar)
	isPercent = isTokenType(TokenPercent)
	isDot     = isTokenType(TokenDot)

	isBackslash = isTokenType(TokenBackslash)
)

// New initializes a Lexer object
func New(r io.Reader) *Lexer {
	s := &scanner.Scanner{
		Mode: scanner.ScanIdents | scanner.ScanFloats | scanner.ScanChars | scanner.ScanStrings | scanner.ScanRawStrings | scanner.ScanComments,
	}

	return &Lexer{
		in:     s.Init(r),
		tokens: make(chan Token),
		done:   make(chan struct{}),
		buf:    []rune{},
	}
}

// Lexer represents a lexical analyzer
type Lexer struct {
	in *scanner.Scanner

	tokens chan Token

	done    chan struct{}
	lastErr error

	buf []rune

	start  int
	offset int
	lines  int
}

// Tokens returns a channel that is going to receive tokens as soon as they are
// detected.
func (lx *Lexer) Tokens() chan Token {
	return lx.tokens
}

func (lx *Lexer) stop() {
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

// Scan starts scanning the reader for tokens.
func (lx *Lexer) Scan() error {
	for state := lexDefaultState; state != nil; {
		select {
		case <-lx.done:
			return nil
		default:
			state = state(lx)
		}
	}

	if lx.lastErr == nil {
		lx.emit(TokenEOF)
	}

	close(lx.tokens)

	return lx.lastErr
}

func (lx *Lexer) emit(tt TokenType) {
	lx.tokens <- Token{
		tt:   tt,
		text: string(lx.buf),

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

	case isQuote(r):
		return lexEmit(TokenQuote)
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
	case isStar(r):
		return lexEmit(TokenStar)
	case isPercent(r):
		return lexEmit(TokenPercent)
	case isDot(r):
		return lexEmit(TokenDot)
	case isBackslash(r):
		return lexEmit(TokenBackslash)

	default:
		return lexString

	}

	panic("unreachable")
}

func lexString(lx *Lexer) lexState {
	for {
		p := lx.peek()
		if isWhitespace(p) || isNewLine(p) || isQuote(p) {
			break
		}
		if _, err := lx.next(); err != nil {
			return lexStateError(err)
		}
	}
	lx.emit(TokenString)
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
	return func(lx *Lexer) lexState {
		log.Printf("lexer error: %v", err)
		lx.lastErr = err
		return nil
	}
}

func lexStateEOF(lx *Lexer) lexState {
	return nil
}

// TokenizeBytes takes an array of bytes and returns all the tokens within it,
// or an error if a token can't be identified.
func TokenizeBytes(in []byte) ([]Token, error) {
	tokens := []Token{}
	done := make(chan struct{})

	lx := New(bytes.NewReader(in))

	go func() {
		for tok := range lx.tokens {
			tokens = append(tokens, tok)
		}
		done <- struct{}{}
	}()

	if err := lx.Scan(); err != nil {
		return nil, err
	}

	<-done
	return tokens, nil
}
