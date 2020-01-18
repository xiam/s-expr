package lexer

import (
	"fmt"
)

// Token represents a known sequence of characters (lexical unit)
type Token struct {
	tt     TokenType
	lexeme string

	line int
	col  int
}

// NewToken creates a lexical unit
func NewToken(tt TokenType, lexeme string, line int, col int) *Token {
	return &Token{
		tt:     tt,
		lexeme: lexeme,
		line:   line,
		col:    col,
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
	return t.lexeme
}

// Is returns true if the token matches the given type
func (t Token) Is(tt TokenType) bool {
	return t.tt == tt
}

func (t Token) String() string {
	return fmt.Sprintf("(:%v %q [%d %d])", tokenName(t.tt), t.lexeme, t.line, t.col)
}
