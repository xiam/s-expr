package sexpr

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/xiam/sexpr/lexer"
)

var (
	errUnexpectedEOF   = errors.New("unexpected EOF")
	errUnexpectedToken = errors.New("unexpected token")
)

var TokenEOF = lexer.NewToken(lexer.TokenEOF, "", 0, 0)

type parserState func(p *parser) parserState

type NodeType uint8

const (
	NodeTypeExpression NodeType = iota
	NodeTypeList
	NodeTypeAtom
	NodeTypeMap
)

var nodeTypeName = map[NodeType]string{
	NodeTypeExpression: "expression",
	NodeTypeList:       "list",
	NodeTypeMap:        "map",
	NodeTypeAtom:       "atom",
}

type Node struct {
	Type     NodeType
	Value    Value
	Children []*Node

	token *lexer.Token
}

func NewNode(tok *lexer.Token, value interface{}) (*Node, error) {
	v, err := NewValue(value)
	if err != nil {
		return nil, err
	}

	return &Node{
		Type:  NodeTypeAtom,
		Value: *v,
		token: tok,
	}, nil
}

func NewExpressionNode(tok *lexer.Token) *Node {
	return &Node{
		Type:     NodeTypeExpression,
		Children: []*Node{},

		token: tok,
	}
}

func NewMapNode(tok *lexer.Token) *Node {
	return &Node{
		Type:     NodeTypeMap,
		Children: []*Node{},

		token: tok,
	}
}

func NewListNode(tok *lexer.Token) *Node {
	return &Node{
		Type:     NodeTypeList,
		Children: []*Node{},

		token: tok,
	}
}

func (n *Node) push(node *Node) {
	n.Children = append(n.Children, node)
}

func (n *Node) NewExpression(tok *lexer.Token) *Node {
	node := NewExpressionNode(tok)
	n.Children = append(n.Children, node)
	return node
}

func (n *Node) NewList(tok *lexer.Token) *Node {
	node := NewListNode(tok)
	n.Children = append(n.Children, node)
	return node
}

func (n *Node) NewMap(tok *lexer.Token) *Node {
	node := NewMapNode(tok)
	n.Children = append(n.Children, node)
	return node
}

func (n Node) String() string {
	switch n.Type {
	case NodeTypeExpression, NodeTypeList, NodeTypeMap:
		return fmt.Sprintf("(%v)[%d]", nodeTypeName[n.Type], len(n.Children))
	}
	return fmt.Sprintf("(%v): %v", nodeTypeName[n.Type], n.Value)
}

type parser struct {
	lx   *lexer.Lexer
	root *Node

	lastTok *lexer.Token
	nextTok *lexer.Token

	lastErr error
}

func newParser(r io.Reader) *parser {
	p := &parser{}
	p.root = &Node{
		Type:     NodeTypeList,
		Children: []*Node{},
	}
	p.lx = lexer.New(r)
	return p
}

func (p *parser) run() error {
	//errCh := make(chan error)

	go func() {
		err := p.lx.Scan()
		_ = err
		//errCh <- err
	}()

	for state := parserDefaultState(p); state != nil; {
		state = state(p)
	}

	//p.lx.stop()
	/*
		err := <-errCh
		if err != nil {
			return err
		}
	*/

	return p.lastErr
}

func (p *parser) curr() *lexer.Token {
	return p.lastTok
}

func (p *parser) read() *lexer.Token {
	tok, ok := <-p.lx.Tokens()
	if ok {
		return &tok
	}
	return TokenEOF
}

func (p *parser) peek() *lexer.Token {
	if p.nextTok != nil {
		return p.nextTok
	}

	p.nextTok = p.read()
	return p.nextTok
}

func (p *parser) next() *lexer.Token {
	if p.nextTok != nil {
		tok := p.nextTok
		p.lastTok, p.nextTok = tok, nil
		return tok
	}

	tok := p.read()
	p.lastTok, p.nextTok = tok, nil
	return tok
}

func parserDefaultState(p *parser) parserState {
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
	return func(p *parser) parserState {
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

func expectTokens(p *parser, tt ...lexer.TokenType) ([]*lexer.Token, error) {
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

func parserStateData(root *Node) parserState {
	return func(p *parser) parserState {
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
			if state := parserStateAtom(root)(p); state != nil {
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
			if state := parserStateOpenMap(root.NewMap(tok))(p); state != nil {
				return state
			}

		case lexer.TokenOpenList:
			if state := parserStateOpenList(root.NewList(tok))(p); state != nil {
				return state
			}

		case lexer.TokenOpenExpression:
			if state := parserStateOpenExpression(root.NewExpression(tok))(p); state != nil {
				return state
			}

		default:
			return parserErrorState(errUnexpectedToken)

		}

		return nil
	}
}

func expectIntegerNode(p *parser) (*Node, error) {
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

		return NewNode(tok, f64)

	default:
		// natural end for an integer
		i64, err := strconv.ParseInt(curr.Text(), 10, 64)
		if err != nil {
			return nil, err
		}

		return NewNode(curr, i64)
	}

	panic("unreachable")
}

func expectString(p *parser) (*lexer.Token, error) {
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

func expectComment(p *parser) (string, error) {
	for {
		tok := p.next()
		switch tok.Type() {
		case lexer.TokenNewLine, lexer.TokenEOF:
			return "", nil
		}
	}
}

func parserStateComment(root *Node) parserState {
	return func(p *parser) parserState {
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

func parserStateString(root *Node) parserState {
	return func(p *parser) parserState {
		tokens := []*lexer.Token{}

	loop:
		for {
			tok := p.next()

			switch tok.Type() {
			case lexer.TokenQuote:
				break loop

			case lexer.TokenEOF:
				return parserErrorState(errUnexpectedEOF)

			default:
				tokens = append(tokens, tok)
			}
		}

		tok := mergeTokens(lexer.TokenString, tokens)

		node, err := NewNode(tok, tok.Text())
		if err != nil {
			return parserErrorState(errUnexpectedEOF)
		}

		root.push(node)
		return nil
	}
}

func parserStateArgument(root *Node) parserState {
	return func(p *parser) parserState {
		curr := p.curr()

		val := p.next()
		switch val.Type() {
		case lexer.TokenInteger, lexer.TokenStar:
			// ok
		default:
			return parserErrorState(errUnexpectedToken)
		}

		tok := mergeTokens(lexer.TokenLiteral, append([]*lexer.Token{curr}, val))
		node, err := NewNode(tok, tok.Text())
		if err != nil {
			return parserErrorState(err)
		}
		root.push(node)
		return nil
	}
}

func parserStateNumeric(root *Node) parserState {
	return func(p *parser) parserState {
		node, err := expectIntegerNode(p)
		if err != nil {
			return parserErrorState(err)
		}
		root.push(node)
		return nil
	}
}

func parserStateWord(root *Node) parserState {
	return func(p *parser) parserState {
		curr := p.curr()

		node, err := NewNode(curr, curr.Text())
		if err != nil {
			return parserErrorState(err)
		}
		root.push(node)

		return nil
	}
}

func parserStateAtom(root *Node) parserState {
	return func(p *parser) parserState {
		curr := p.curr()

		atomName, err := expectTokens(p, lexer.TokenWord)
		if err != nil {
			return parserErrorState(err)
		}

		tok := mergeTokens(lexer.TokenLiteral, append([]*lexer.Token{curr}, atomName...))
		node, err := NewNode(tok, tok.Text())
		if err != nil {
			return parserErrorState(err)
		}
		node.Value.Type = ValueTypeAtom
		root.push(node)
		return nil
	}
}

func parserStateOpenMap(root *Node) parserState {
	return func(p *parser) parserState {
		tok := p.next()

		switch tok.Type() {
		case lexer.TokenEOF:
			return parserErrorState(errUnexpectedEOF)

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

func parserStateOpenExpression(root *Node) parserState {
	return func(p *parser) parserState {
		tok := p.next()

		switch tok.Type() {
		case lexer.TokenEOF:
			return parserErrorState(errUnexpectedEOF)

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

func parserStateOpenList(root *Node) parserState {
	return func(p *parser) parserState {
		tok := p.next()

		switch tok.Type() {
		case lexer.TokenEOF:
			return parserErrorState(errUnexpectedEOF)

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

func parse(in []byte) (*Node, error) {
	p := newParser(bytes.NewReader(in))

	err := p.run()
	if err != nil {
		return nil, err
	}

	return p.root, nil
}

func printNode(node *Node) {
	printNodeLevel(node, 0)
}

func compileNode(node *Node) []byte {
	return compileNodeLevel(node, 0)
}

func compileNodeLevel(node *Node, level int) []byte {
	if node == nil {
		return []byte(":nil")
	}
	switch node.Type {
	case NodeTypeMap:
		nodes := []string{}
		for i := range node.Children {
			nodes = append(nodes, string(compileNodeLevel(node.Children[i], level+1)))
		}
		return []byte(fmt.Sprintf("{%s}", strings.Join(nodes, " ")))

	case NodeTypeList:
		nodes := []string{}
		for i := range node.Children {
			nodes = append(nodes, string(compileNodeLevel(node.Children[i], level+1)))
		}
		return []byte(fmt.Sprintf("[%s]", strings.Join(nodes, " ")))

	case NodeTypeExpression:
		nodes := []string{}
		for i := range node.Children {
			nodes = append(nodes, string(compileNodeLevel(node.Children[i], level+1)))
		}
		if level == 0 {
			return []byte(fmt.Sprintf("%s", strings.Join(nodes, " ")))
		}
		return []byte(fmt.Sprintf("(%s)", strings.Join(nodes, " ")))

	case NodeTypeAtom:
		if node.token.Is(lexer.TokenString) {
			return []byte(fmt.Sprintf("%q", node.Value))
		}
		return []byte(fmt.Sprintf("%v", node.Value))

	default:
		panic("unknown node type")
	}
}

func printNodeLevel(node *Node, level int) {
	if node == nil {
		fmt.Printf(":nil\n")
		return
	}
	indent := strings.Repeat("    ", level)
	fmt.Printf("%s(%s): ", indent, nodeTypeName[node.Type])
	switch node.Type {

	case NodeTypeExpression, NodeTypeList, NodeTypeMap:
		fmt.Printf("(%v)\n", node.token)
		for i := range node.Children {
			printNodeLevel(node.Children[i], level+1)
		}

	case NodeTypeAtom:
		fmt.Printf("%#v (%v)\n", node.Value, node.token)

	default:
		panic("unknown node type")
	}
}

func parserError(err error, tok *lexer.Token) error {
	log.Fatalf("%v: %v", err.Error(), tok)
	return err
}
