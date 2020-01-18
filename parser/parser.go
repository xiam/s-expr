package parser

import (
	"bytes"
	"errors"
	"io"
	"log"
	"strconv"

	"github.com/xiam/sexpr/ast"
	"github.com/xiam/sexpr/lexer"
)

var (
	errUnexpectedEOF   = errors.New("unexpected EOF")
	errUnexpectedToken = errors.New("unexpected token")
)

var TokenEOF = lexer.NewToken(lexer.TokenEOF, "", 0, 0)

type parserState func(p *Parser) parserState

type Parser struct {
	lx   *lexer.Lexer
	root *ast.Node

	lastTok *lexer.Token
	nextTok *lexer.Token

	lastErr error
}

func New(r io.Reader) *Parser {
	p := &Parser{}
	p.root = ast.NewList(nil)
	p.lx = lexer.New(r)
	return p
}

func (p *Parser) Parse() error {
	errCh := make(chan error)

	go func() {
		errCh <- p.lx.Scan()
	}()

	for state := parserDefaultState(p); state != nil; {
		state = state(p)
	}

	err := <-errCh
	if err != nil {
		return err
	}

	return p.lastErr
}

func (p *Parser) curr() *lexer.Token {
	return p.lastTok
}

func (p *Parser) read() *lexer.Token {
	tok, ok := <-p.lx.Tokens()
	if ok {
		return &tok
	}
	return TokenEOF
}

func (p *Parser) peek() *lexer.Token {
	if p.nextTok != nil {
		return p.nextTok
	}

	p.nextTok = p.read()
	return p.nextTok
}

func (p *Parser) next() *lexer.Token {
	if p.nextTok != nil {
		tok := p.nextTok
		p.lastTok, p.nextTok = tok, nil
		return tok
	}

	tok := p.read()
	p.lastTok, p.nextTok = tok, nil
	return tok
}

func parserDefaultState(p *Parser) parserState {
	root := p.root
	tok := p.next()

	switch tok.Type() {
	case lexer.TokenEOF:
		return nil

	default:
		if state := parserStateData(root)(p); state != nil {
			return state
		}
	}

	return parserDefaultState
}

func ParserErrorState(err error) parserState {
	return func(p *Parser) parserState {
		//p.lx.stop()
		p.lastErr = err
		return nil
	}
}

func mergeTokens(tt lexer.TokenType, tokens []*lexer.Token) *lexer.Token {
	var text string

	var firstTok *lexer.Token
	for _, tok := range tokens {
		if firstTok == nil {
			firstTok = tok
		}
		text = text + tok.Text()
	}

	line, col := firstTok.Pos()
	return lexer.NewToken(tt, text, line, col)
}

func expectTokens(p *Parser, tt ...lexer.TokenType) ([]*lexer.Token, error) {
	tokens := []*lexer.Token{}
	for i := range tt {
		tok := p.next()
		if tok.Type() == lexer.TokenEOF {
			return nil, errUnexpectedEOF
		}
		if tok.Type() != tt[i] {
			return nil, errUnexpectedToken
		}
		tokens = append(tokens, tok)
	}
	return tokens, nil
}

func parserStateData(root *ast.Node) parserState {
	return func(p *Parser) parserState {
		tok := p.curr()

		switch tok.Type() {
		case lexer.TokenWhitespace, lexer.TokenNewLine:
			// continue

		case lexer.TokenQuote:
			if state := parserStateString(root)(p); state != nil {
				return state
			}

		case lexer.TokenInteger:
			if state := parserStateNumeric(root)(p); state != nil {
				return state
			}

		case lexer.TokenPercent:
			if state := parserStateArgument(root)(p); state != nil {
				return state
			}

		case lexer.TokenColon:
			if state := parserStateValue(root)(p); state != nil {
				return state
			}

		case lexer.TokenWord:
			if state := parserStateWord(root)(p); state != nil {
				return state
			}

		case lexer.TokenString:
			if state := parserStateWord(root)(p); state != nil {
				return state
			}

		case lexer.TokenHash:
			if state := parserStateComment(root)(p); state != nil {
				return state
			}

		case lexer.TokenOpenMap:
			if state := parserStateOpenMap(root.PushMap(tok))(p); state != nil {
				return state
			}

		case lexer.TokenOpenList:
			if state := parserStateOpenList(root.PushList(tok))(p); state != nil {
				return state
			}

		case lexer.TokenOpenExpression:
			if state := parserStateOpenExpression(root.PushExpression(tok))(p); state != nil {
				return state
			}

		default:
			return ParserErrorState(errUnexpectedToken)
		}

		return nil
	}
}

func expectIntegerNode(p *Parser) (*ast.Node, error) {
	curr := p.curr()

	next := p.peek()
	switch next.Type() {
	case lexer.TokenDot:
		// got a point, this means this is a floating point number

		mantissa, err := expectTokens(p, lexer.TokenDot, lexer.TokenInteger)
		if err != nil {
			return nil, err
		}

		tok := mergeTokens(lexer.TokenLiteral, append([]*lexer.Token{curr}, mantissa...))

		f64, err := strconv.ParseFloat(tok.Text(), 64)
		if err != nil {
			return nil, err
		}

		return ast.New(tok, f64), nil

	default:
		// natural end for an integer
		i64, err := strconv.ParseInt(curr.Text(), 10, 64)
		if err != nil {
			return nil, err
		}

		return ast.New(curr, i64), nil
	}

	panic("unreachable")
}

func expectString(p *Parser) (*lexer.Token, error) {
	tokens := []*lexer.Token{}

loop:
	for {
		tok := p.next()

		switch tok.Type() {
		case lexer.TokenQuote:
			break loop
		case lexer.TokenEOF:
			return nil, errUnexpectedEOF
		default:
			tokens = append(tokens, tok)
		}
	}

	return mergeTokens(lexer.TokenLiteral, tokens), nil
}

func expectComment(p *Parser) (string, error) {
	for {
		tok := p.next()
		switch tok.Type() {
		case lexer.TokenNewLine, lexer.TokenEOF:
			return "", nil
		}
	}
}

func parserStateComment(root *ast.Node) parserState {
	return func(p *Parser) parserState {
	loop:
		for {
			tok := p.next()
			switch tok.Type() {
			case lexer.TokenEOF, lexer.TokenNewLine:
				break loop
			}
		}
		return nil
	}
}

func parserStateString(root *ast.Node) parserState {
	return func(p *Parser) parserState {
		tokens := []*lexer.Token{}

	loop:
		for {
			tok := p.next()

			switch tok.Type() {
			case lexer.TokenQuote:
				break loop

			case lexer.TokenEOF:
				return ParserErrorState(errUnexpectedEOF)

			default:
				tokens = append(tokens, tok)
			}
		}

		tok := mergeTokens(lexer.TokenString, tokens)
		_ = ast.New(tok, tok.Text())
		return nil
	}
}

func parserStateArgument(root *ast.Node) parserState {
	return func(p *Parser) parserState {
		curr := p.curr()

		val := p.next()
		switch val.Type() {
		case lexer.TokenInteger, lexer.TokenStar:
			// ok
		default:
			return ParserErrorState(errUnexpectedToken)
		}

		tok := mergeTokens(lexer.TokenLiteral, append([]*lexer.Token{curr}, val))
		_ = ast.New(tok, tok.Text())
		return nil
	}
}

func parserStateNumeric(root *ast.Node) parserState {
	return func(p *Parser) parserState {
		node, err := expectIntegerNode(p)
		if err != nil {
			return ParserErrorState(err)
		}
		root.AppendChild(node)
		return nil
	}
}

func parserStateWord(root *ast.Node) parserState {
	return func(p *Parser) parserState {
		curr := p.curr()
		root.Push(curr, curr.Text())
		return nil
	}
}

func parserStateValue(root *ast.Node) parserState {
	return func(p *Parser) parserState {
		curr := p.curr()

		atomName, err := expectTokens(p, lexer.TokenWord)
		if err != nil {
			return ParserErrorState(err)
		}

		tok := mergeTokens(lexer.TokenLiteral, append([]*lexer.Token{curr}, atomName...))
		node := ast.New(tok, tok.Text())
		//ast.Value.Type = ValueTypeValue
		root.AppendChild(node)
		return nil
	}
}

func parserStateOpenMap(root *ast.Node) parserState {
	return func(p *Parser) parserState {
		tok := p.next()

		switch tok.Type() {
		case lexer.TokenEOF:
			return ParserErrorState(errUnexpectedEOF)

		case lexer.TokenCloseMap:
			return nil

		default:
			if state := parserStateData(root)(p); state != nil {
				return state
			}
		}

		return parserStateOpenMap(root)(p)
	}
}

func parserStateOpenExpression(root *ast.Node) parserState {
	return func(p *Parser) parserState {
		tok := p.next()

		switch tok.Type() {
		case lexer.TokenEOF:
			return ParserErrorState(errUnexpectedEOF)

		case lexer.TokenCloseExpression:
			return nil

		default:
			if state := parserStateData(root)(p); state != nil {
				return state
			}
		}

		return parserStateOpenExpression(root)(p)
	}
}

func parserStateOpenList(root *ast.Node) parserState {
	return func(p *Parser) parserState {
		tok := p.next()

		switch tok.Type() {
		case lexer.TokenEOF:
			return ParserErrorState(errUnexpectedEOF)

		case lexer.TokenCloseList:
			return nil

		default:
			if state := parserStateData(root)(p); state != nil {
				return state
			}
		}

		return parserStateOpenList(root)(p)
	}
}

func Parse(in []byte) (*ast.Node, error) {
	p := New(bytes.NewReader(in))

	err := p.Parse()
	if err != nil {
		return nil, err
	}

	return p.root, nil
}

func ParserError(err error, tok *lexer.Token) error {
	log.Fatalf("%v: %v", err.Error(), tok)
	return err
}
