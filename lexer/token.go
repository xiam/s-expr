package lexer

import (
	"fmt"
)

// EOF represents the end of a file
var EOF = NewToken(TokenEOF, "", -1, -1)

// Token represents a known sequence of characters (lexical unit)
type Token struct {
	tt   TokenType
	text string

	line int
	col  int
}

// NewToken creates a lexical unit
func NewToken(tt TokenType, text string, line int, col int) *Token {
	return &Token{
		tt:   tt,
		text: text,
		line: line,
		col:  col,
	}
}

// Type returns the type of the lexical unit
func (t Token) Type() TokenType {
	return t.tt
}

// Pos returns the line and column of the lexical unit
func (t Token) Pos() (int, int) {
	return t.line, t.col
}

// Text returns the raw text of the lexical unit
func (t Token) Text() string {
	return t.text
}

// Is returns true if the token matches the given type
func (t Token) Is(tt TokenType) bool {
	return t.tt == tt
}

func (t Token) String() string {
	return fmt.Sprintf("%q (%v) %d:%d", t.text, tokenName(t.tt), t.line, t.col)
}
