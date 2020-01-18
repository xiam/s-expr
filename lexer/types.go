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
	TokenDoubleQuote
	TokenHash
	TokenWhitespace

	TokenWord
	TokenInteger
	TokenSequence

	TokenColon
	TokenDot
	TokenBackslash

	TokenEOF
)

var tokenValues = map[TokenType][]rune{
	TokenOpenList:        []rune{'['},
	TokenCloseList:       []rune{']'},
	TokenOpenMap:         []rune{'{'},
	TokenCloseMap:        []rune{'}'},
	TokenOpenExpression:  []rune{'('},
	TokenCloseExpression: []rune{')'},
	TokenNewLine:         []rune{'\n'},
	TokenDoubleQuote:     []rune{'"'},
	TokenHash:            []rune{'#'},
	TokenWhitespace:      []rune(" \f\t\r"),
	TokenWord:            []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_"),
	TokenInteger:         []rune("0123456789"),
	TokenColon:           []rune{':'},
	TokenDot:             []rune{'.'},
	TokenBackslash:       []rune{'\\'},
}

var tokenNames = map[TokenType]string{
	TokenInvalid:         "invalid",
	TokenOpenList:        "open_list",
	TokenCloseList:       "close_list",
	TokenOpenMap:         "open_map",
	TokenCloseMap:        "close_map",
	TokenOpenExpression:  "open_expression",
	TokenCloseExpression: "close_expression",
	TokenNewLine:         "newline",
	TokenDoubleQuote:     "double_quote",
	TokenHash:            "hash",
	TokenWhitespace:      "separator",
	TokenWord:            "word",
	TokenInteger:         "integer",
	TokenColon:           "colon",
	TokenDot:             "dot",
	TokenSequence:        "sequence",
	TokenEOF:             "EOF",
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
