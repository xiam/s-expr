package lexer

// TokenType represents all the possible types of a lexical unit
type TokenType uint8

// List of types of lexical units
const (
	TokenInvalid         TokenType = iota
	TokenOpenExpression            // Open parenthesis: "["
	TokenCloseExpression           // Close parenthesis: "]"
	TokenOpenList                  // Open square bracket: "("
	TokenCloseList                 // Close square bracker: ")"
	TokenOpenMap                   // Open curly bracket: "{"
	TokenCloseMap                  // Close curly bracker: "}"
	TokenNewLine                   // Newline: "\n"
	TokenDoubleQuote               // Double quote: '"'
	TokenHash                      // Hash: "#"
	TokenWhitespace                // Space, tab, linefeed or carriage return: \s\f\t\r
	TokenWord                      // Letters ([a-zA-Z]) and underscore
	TokenInteger                   // Integers
	TokenSequence                  // Extended sequence
	TokenColon                     // Colon: ":"
	TokenDot                       // Dot: "."
	TokenBackslash                 // Backslash: "\"
	TokenEOF                       // End of file
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

func (tt TokenType) String() string {
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

func isAritmeticSign(p rune) bool {
	return p == '+' || p == '-'
}
