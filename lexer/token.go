package lexer

import (
	"fmt"
	"text/scanner"
)

// Token represents a known sequence of characters (lexical unit)
type Token struct {
	tt     TokenType
	lexeme string

	pos scanner.Position
}

// NewToken creates a lexical unit
func NewToken(tt TokenType, lexeme string, pos *scanner.Position) *Token {
	if pos == nil {
		pos = &scanner.Position{}
	}
	return &Token{
		tt:     tt,
		lexeme: lexeme,
		pos:    *pos,
	}
}

// Type returns the type of the lexical unit
func (t Token) Type() TokenType {
	return t.tt
}

// Pos returns the line and column of the lexical unit
func (t Token) Pos() scanner.Position {
	return t.pos
}

// Text returns the raw text of the lexical unit
func (t Token) Text() string {
	return t.lexeme
}

func (t Token) String() string {
	return fmt.Sprintf("(:%v %q [%d %d])", t.tt, t.lexeme, t.pos.Line, t.pos.Column)
}
