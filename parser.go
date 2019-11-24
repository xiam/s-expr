package sexpr

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

var (
	errUnexpectedEOF   = errors.New("unexpected EOF")
	errUnexpectedToken = errors.New("unexpected token")
)

var TokenEOF = token{tt: tokenEOF}

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
	Value    interface{}
	Children []*Node

	token *token

	child chan *Node
}

func NewNode(tok *token, value interface{}) (*Node, error) {
	value, err := NewValue(value)
	if err != nil {
		return nil, err
	}

	return &Node{
		Type:  NodeTypeAtom,
		Value: value,
		token: tok,
	}, nil
}

func NewExpressionNode(tok *token) *Node {
	return &Node{
		Type:     NodeTypeExpression,
		Children: []*Node{},

		token: tok,
	}
}

func NewMapNode(tok *token) *Node {
	return &Node{
		Type:     NodeTypeMap,
		Children: []*Node{},

		token: tok,
	}
}

func NewListNode(tok *token) *Node {
	return &Node{
		Type:     NodeTypeList,
		Children: []*Node{},

		token: tok,
	}
}

func (n *Node) push(node *Node) {
	n.Children = append(n.Children, node)
}

func (n *Node) NewExpression(tok *token) *Node {
	node := NewExpressionNode(tok)
	n.Children = append(n.Children, node)
	return node
}

func (n *Node) NewList(tok *token) *Node {
	node := NewListNode(tok)
	n.Children = append(n.Children, node)
	return node
}

func (n *Node) NewMap(tok *token) *Node {
	node := NewMapNode(tok)
	n.Children = append(n.Children, node)
	return node
}

func (n *Node) Serve() {
	n.child = make(chan *Node)

	go func() {
		defer n.Close()
		for i := range n.Children {
			n.child <- n.Children[i]
		}
	}()
}

func (n *Node) Close() {
	close(n.child)
}

func (n *Node) Next() *Node {
	node, ok := <-n.child
	if !ok {
		return nil
	}
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
	lx   *lexer
	root *Node

	lastTok *token
	nextTok *token

	lastErr error
}

func newParser(r io.Reader) *parser {
	p := &parser{}
	p.root = &Node{
		Type:     NodeTypeExpression,
		Children: []*Node{},
	}
	p.lx = newLexer(r)
	return p
}

func (p *parser) run() error {
	//errCh := make(chan error)

	go func() {
		err := p.lx.run()
		log.Printf("ERR: %v", err)
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

func (p *parser) curr() *token {
	return p.lastTok
}

func (p *parser) read() *token {
	tok, ok := <-p.lx.tokens
	if ok {
		return &tok
	}
	return &TokenEOF
}

func (p *parser) peek() *token {
	if p.nextTok != nil {
		return p.nextTok
	}

	p.nextTok = p.read()
	return p.nextTok
}

func (p *parser) next() *token {
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

	switch tok.tt {
	case tokenEOF:
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
		log.Printf("err: %v -- %v", err, p.lastTok)
		return nil
	}
}

func mergeTokens(tokens []*token) *token {
	var firstTok *token
	var text string

	tt := tokenInvalid
	for _, tok := range tokens {
		if firstTok == nil {
			firstTok = tok
			tt = tok.tt
		}
		if tt != tok.tt {
			tt = tokenLiteral
		}
		text = text + tok.text
	}

	return &token{
		tt:   tt,
		text: text,

		col:  firstTok.col,
		line: firstTok.line,
	}
}

func expectTokens(p *parser, tt ...tokenType) ([]*token, error) {
	tokens := []*token{}
	for i := range tt {
		tok := p.next()
		if tok.tt == tokenEOF {
			return nil, errUnexpectedEOF
		}
		if tok.tt != tt[i] {
			return nil, errUnexpectedToken
		}
		tokens = append(tokens, tok)
	}
	return tokens, nil
}

func parserStateData(root *Node) parserState {
	return func(p *parser) parserState {
		tok := p.curr()

		switch tok.tt {
		case tokenWhitespace, tokenNewLine:
			// continue

		case tokenQuote:
			if state := parserStateString(root)(p); state != nil {
				return state
			}

		case tokenInteger:
			if state := parserStateNumeric(root)(p); state != nil {
				return state
			}

		case tokenPercent:
			if state := parserStateArgument(root)(p); state != nil {
				return state
			}

		case tokenColon:
			if state := parserStateAtom(root)(p); state != nil {
				return state
			}

		case tokenWord:
			if state := parserStateWord(root)(p); state != nil {
				return state
			}

		case tokenHash:
			if state := parserStateComment(root)(p); state != nil {
				return state
			}

		case tokenOpenMap:
			if state := parserStateOpenMap(root.NewMap(tok))(p); state != nil {
				return state
			}

		case tokenOpenList:
			if state := parserStateOpenList(root.NewList(tok))(p); state != nil {
				return state
			}

		case tokenOpenExpression:
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
	switch next.tt {
	case tokenDot:
		// got a point, this means this is a floating point number

		mantissa, err := expectTokens(p, tokenDot, tokenInteger)
		if err != nil {
			return nil, err
		}

		tok := mergeTokens(append([]*token{curr}, mantissa...))

		f64, err := strconv.ParseFloat(tok.text, 64)
		if err != nil {
			return nil, err
		}

		return NewNode(tok, f64)

	default:
		// natural end for an integer
		i64, err := strconv.ParseInt(curr.text, 10, 64)
		if err != nil {
			return nil, err
		}

		return NewNode(curr, i64)
	}

	panic("unreachable")
}

func expectString(p *parser) (*token, error) {
	tokens := []*token{}

loop:
	for {
		tok := p.next()

		switch tok.tt {
		case tokenQuote:
			break loop
		case tokenEOF:
			return nil, errUnexpectedEOF
		default:
			tokens = append(tokens, tok)
		}
	}

	return mergeTokens(tokens), nil
}

func expectComment(p *parser) (string, error) {
	for {
		tok := p.next()
		switch tok.tt {
		case tokenNewLine, tokenEOF:
			return "", nil
		}
	}
}

func parserStateComment(root *Node) parserState {
	return func(p *parser) parserState {
	loop:
		for {
			tok := p.next()
			switch tok.tt {
			case tokenEOF, tokenNewLine:
				break loop
			}
		}
		return nil
	}
}

func parserStateString(root *Node) parserState {
	return func(p *parser) parserState {
		tokens := []*token{}

	loop:
		for {
			tok := p.next()
			switch tok.tt {
			case tokenQuote:
				break loop

			case tokenEOF:
				return parserErrorState(errUnexpectedEOF)

			default:
				tokens = append(tokens, tok)
			}
		}

		tok := mergeTokens(tokens)
		tok.tt = tokenString

		node, err := NewNode(tok, tok.text)
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
		switch val.tt {
		case tokenInteger, tokenStar:
			// ok
		default:
			return parserErrorState(errUnexpectedToken)
		}

		tok := mergeTokens(append([]*token{curr}, val))
		node, err := NewNode(tok, tok.text)
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

		node, err := NewNode(curr, curr.text)
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

		atomName, err := expectTokens(p, tokenWord)
		if err != nil {
			return parserErrorState(err)
		}

		tok := mergeTokens(append([]*token{curr}, atomName...))
		node, err := NewNode(tok, tok.text)
		if err != nil {
			return parserErrorState(err)
		}
		root.push(node)
		return nil
	}
}

func parserStateOpenMap(root *Node) parserState {
	return func(p *parser) parserState {
		tok := p.next()

		switch tok.tt {
		case tokenEOF:
			return parserErrorState(errUnexpectedEOF)

		default:
			if state := parserStateData(root)(p); state != nil {
				return nil
			}
		}

		return parserStateOpenMap(root)(p)
	}
}

func parserStateOpenExpression(root *Node) parserState {
	return func(p *parser) parserState {
		tok := p.next()

		switch tok.tt {
		case tokenEOF:
			return parserErrorState(errUnexpectedEOF)

		default:
			if state := parserStateData(root)(p); state != nil {
				return nil
			}
		}

		return parserStateOpenExpression(root)(p)
	}
}

func parserStateOpenList(root *Node) parserState {
	return func(p *parser) parserState {
		tok := p.next()

		switch tok.tt {
		case tokenEOF:
			return parserErrorState(errUnexpectedEOF)

		default:
			if state := parserStateData(root)(p); state != nil {
				return nil
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
		if node.token.tt == tokenString {
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

func parserError(err error, tok *token) error {
	return fmt.Errorf("%v: %v", err.Error(), tok)
}
