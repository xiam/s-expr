package lexer

// TokenType represents all the possible types of a lexical unit
type TokenType uint8

// List of types of lexical units
const (
	TokenInvalid TokenType = iota

	TokenOpenList
	TokenCloseList

	TokenOpenMap
	TokenCloseMap

	TokenOpenExpression
	TokenCloseExpression
	TokenExpression

	TokenNewLine
	TokenQuote
	TokenHash
	TokenWhitespace

	TokenWord
	TokenInteger
	TokenString
	TokenLiteral

	TokenColon
	TokenStar
	TokenPercent
	TokenDot

	TokenBackslash

	TokenEOF
)

var tokenValues = map[TokenType][]rune{
	TokenOpenList:  []rune{'['},
	TokenCloseList: []rune{']'},

	TokenOpenMap:  []rune{'{'},
	TokenCloseMap: []rune{'}'},

	TokenOpenExpression:  []rune{'('},
	TokenCloseExpression: []rune{')'},

	TokenNewLine:    []rune{'\n'},
	TokenQuote:      []rune{'"'},
	TokenHash:       []rune{'#'},
	TokenWhitespace: []rune(" \f\t\r"),

	TokenWord:    []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ@_"),
	TokenInteger: []rune("0123456789"),

	TokenColon:   []rune{':'},
	TokenStar:    []rune{'*'},
	TokenPercent: []rune{'%'},
	TokenDot:     []rune{'.'},

	TokenBackslash: []rune{'\\'},
}

var tokenNames = map[TokenType]string{
	TokenInvalid: "[invalid]",

	TokenOpenList:  "[open list]",
	TokenCloseList: "[close list]",

	TokenOpenMap:  "[open map]",
	TokenCloseMap: "[close map]",

	TokenOpenExpression:  "[open expr]",
	TokenCloseExpression: "[close expr]",
	TokenExpression:      "[expression]",

	TokenNewLine:    "[newline]",
	TokenQuote:      "[quote]",
	TokenHash:       "[hash]",
	TokenWhitespace: "[separator]",

	TokenWord:    "[word]",
	TokenInteger: "[integer]",

	TokenColon:   "[colon]",
	TokenStar:    "[star]",
	TokenPercent: "[percent]",
	TokenDot:     "[dot]",
	TokenString:  "[string]",
	TokenLiteral: "[literal]",

	TokenBackslash: "[backslash]",

	TokenEOF: "[EOF]",
}

func tokenName(tt TokenType) string {
	if v, ok := tokenNames[tt]; ok {
		return v
	}
	return tokenNames[TokenInvalid]
}

func isTokenType(tt TokenType) func(r rune) bool {
	return func(r rune) bool {
		for _, v := range tokenValues[tt] {
			if v == r {
				return true
			}
		}
		return false
	}
}
