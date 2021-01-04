package parser

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/xiam/sexpr/ast"
	"github.com/xiam/sexpr/lexer"
)

// EOF represents the end of the file the parser is reading
var EOF = lexer.NewToken(lexer.TokenEOF, "", nil)

type parserState func(p *Parser) parserState

// Parser represents a parser
type Parser struct {
	lx   *lexer.Lexer
	root *ast.Node

	lastTok *lexer.Token
	nextTok *lexer.Token

	options ParserOptions

	lastErr error
}

// New creates a new parser that reads from the given input
func NewParser(r io.Reader) *Parser {
	return &Parser{
		root:    ast.NewList(nil),
		lx:      lexer.New(r),
		options: parserDefaultOptions,
	}
}

func (p *Parser) SetOptions(options ParserOptions) {
	p.options = options
}

func (p *Parser) Options() ParserOptions {
	return p.options
}

func (p *Parser) RootNode() *ast.Node {
	return p.root
}

// Parse tokenizes the input and transforms it into a AST
func (p *Parser) Parse() error {
	errCh := make(chan error)

	go func() {
		errCh <- p.lx.Scan()
	}()

	for state := parserDefaultState(p); state != nil; {
		state = state(p)
	}

	err := <-errCh
	if err != nil && err != lexer.ErrForceStopped {
		return fmt.Errorf("lexer error: %v", err)
	}

	return p.lastErr
}

func (p *Parser) curr() *lexer.Token {
	return p.lastTok
}

func (p *Parser) read() *lexer.Token {
	if ok := p.lx.Next(); !ok {
		return EOF
	}

	return p.lx.Token()
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

func parserErrorState(err error) parserState {
	return func(p *Parser) parserState {
		p.lx.Stop()

		tok := p.curr()
		if tok == nil {
			p.lastErr = fmt.Errorf("syntax error: %w", err)
			return nil
		}

		pos := tok.Pos()
		line, col := pos.Line, pos.Column
		switch err {
		case ErrUnexpectedToken:
			p.lastErr = fmt.Errorf("syntax error: %w %q (around (line %v) (column %v))", err, tok.Text(), line, col)
			return nil
		}
		// TODO: extract a code snippet around line and col
		p.lastErr = fmt.Errorf("syntax error: %w (around (line: %v) (column %v))", err, line, col)
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

	pos := firstTok.Pos()
	return lexer.NewToken(tt, text, &pos)
}

func expectTokens(p *Parser, tt ...lexer.TokenType) ([]*lexer.Token, error) {
	tokens := []*lexer.Token{}
	for i := range tt {
		tok := p.next()
		if tok.Type() == lexer.TokenEOF {
			return nil, ErrUnexpectedEOF
		}
		if tok.Type() != tt[i] {
			return nil, ErrUnexpectedToken
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

		case lexer.TokenDoubleQuote:
			if state := parserStateString(root)(p); state != nil {
				return state
			}

		case lexer.TokenInteger:
			if state := parserStateNumeric(root)(p); state != nil {
				return state
			}

		case lexer.TokenColon:
			if state := parserStateAtom(root)(p); state != nil {
				return state
			}

		case lexer.TokenWord:
			if state := parserStateWord(root)(p); state != nil {
				return state
			}

		case lexer.TokenSequence:
			if state := parserStateWord(root)(p); state != nil {
				return state
			}

		case lexer.TokenHash:
			if state := parserStateComment(root)(p); state != nil {
				return state
			}

		case lexer.TokenOpenMap:
			node, err := root.PushMap(tok)
			if err != nil {
				return parserErrorState(err)
			}
			if state := parserStateOpenMap(node)(p); state != nil {
				return state
			}

		case lexer.TokenOpenList:
			node, err := root.PushList(tok)
			if err != nil {
				return parserErrorState(err)
			}
			if state := parserStateOpenList(node)(p); state != nil {
				return state
			}

		case lexer.TokenOpenExpression:
			node, err := root.PushExpression(tok)
			if err != nil {
				return parserErrorState(err)
			}
			if state := parserStateOpenExpression(node)(p); state != nil {
				return state
			}

		default:
			return parserErrorState(ErrUnexpectedToken)
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

		tok := mergeTokens(lexer.TokenSequence, append([]*lexer.Token{curr}, mantissa...))

		f64, err := strconv.ParseFloat(tok.Text(), 64)
		if err != nil {
			return nil, err
		}

		return ast.NewNode(tok, ast.NewFloatValue(f64)), nil

	default:
		// natural end for an integer
		i64, err := strconv.ParseInt(curr.Text(), 10, 64)
		if err != nil {
			return nil, err
		}

		return ast.NewNode(curr, ast.NewIntValue(i64)), nil
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
			case lexer.TokenDoubleQuote:
				break loop

			case lexer.TokenEOF:
				return parserErrorState(ErrUnexpectedEOF)

			default:
				tokens = append(tokens, tok)
			}
		}

		tok := mergeTokens(lexer.TokenSequence, tokens)
		if err := root.Push(ast.NewNode(tok, ast.NewStringValue(tok.Text()))); err != nil {
			return parserErrorState(err)
		}
		return nil
	}
}

func parserStateNumeric(root *ast.Node) parserState {
	return func(p *Parser) parserState {
		node, err := expectIntegerNode(p)
		if err != nil {
			return parserErrorState(err)
		}
		if err := root.Push(node); err != nil {
			return parserErrorState(err)
		}
		return nil
	}
}

func parserStateWord(root *ast.Node) parserState {
	return func(p *Parser) parserState {
		curr := p.curr()
		if _, err := root.PushValue(curr, ast.NewSymbolValue(curr.Text())); err != nil {
			return parserErrorState(err)
		}
		return nil
	}
}

func parserStateAtom(root *ast.Node) parserState {
	return func(p *Parser) parserState {
		curr := p.curr()

		atomName, err := expectTokens(p, lexer.TokenWord)
		if err != nil {
			return parserErrorState(err)
		}

		tok := mergeTokens(lexer.TokenSequence, append([]*lexer.Token{curr}, atomName...))
		node := ast.NewNode(tok, ast.NewAtomValue(tok.Text()))
		if err := root.Push(node); err != nil {
			return parserErrorState(err)
		}
		return nil
	}
}

func parserStateOpenMap(root *ast.Node) parserState {
	return func(p *Parser) parserState {
		tok := p.next()

		switch tok.Type() {
		case lexer.TokenEOF:
			if p.options.AutoCloseOnEOF {
				return nil
			}
			return parserErrorState(ErrUnexpectedEOF)
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
			if p.options.AutoCloseOnEOF {
				return nil
			}
			return parserErrorState(ErrUnexpectedEOF)

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
			if p.options.AutoCloseOnEOF {
				return nil
			}
			return parserErrorState(ErrUnexpectedEOF)

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

// Parse parses an array of bytes and returns a AST root
func Parse(in []byte) (*ast.Node, error) {
	p := NewParser(bytes.NewReader(in))

	err := p.Parse()
	if err != nil {
		return nil, err
	}

	return p.root, nil
}
