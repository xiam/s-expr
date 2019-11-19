package sexpr

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
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
	NodeTypeList NodeType = iota
	NodeTypeArray
	NodeTypeAtom
)

var nodeTypeName = map[NodeType]string{
	NodeTypeList:  "list",
	NodeTypeArray: "array",
	NodeTypeAtom:  "atom",
}

type Node struct {
	Type     NodeType
	Value    interface{}
	Children []*Node
}

func (n Node) String() string {
	switch n.Type {
	case NodeTypeList, NodeTypeArray:
		return fmt.Sprintf("(%v)[%d]", nodeTypeName[n.Type], len(n.Children))
	}
	return fmt.Sprintf("(%v): %v", nodeTypeName[n.Type], n.Value)
}

type parser struct {
	lx   *lexer
	root *Node
}

func newParser(r io.Reader) *parser {
	p := &parser{}
	p.root = &Node{
		Type:     NodeTypeList,
		Children: []*Node{},
	}
	p.lx = newLexer(r)
	return p
}

func (p *parser) run() error {
	errCh := make(chan error)

	go func() {
		errCh <- p.lx.run()
	}()

	for state := parserDefaultState; state != nil; {
		state = state(p)
	}

	for _ = range p.lx.tokens {

	}

	//p.lx.stop()
	err := <-errCh

	return err
}

func (p *parser) next() *token {
	tok, ok := <-p.lx.tokens
	if ok {
		return &tok
	}
	return &TokenEOF
}

func parserDefaultState(p *parser) parserState {
	tok := p.next()

	switch tok.tt {
	case tokenOpenBracket:
		return parserStateOpenBracket
	case tokenOpenList:
		return parserStateOpenList
	case tokenAtomSeparator, tokenNewLine:
		// ignore
		return parserDefaultState
	case tokenEOF:
		return nil
	default:
		return parserErrorState(errUnexpectedToken)
	}

	panic("unreachable")
}

func parserStateCloseList(p *parser) parserState {
	return parserDefaultState
}

func parserErrorState(err error) parserState {
	return func(p *parser) parserState {
		p.lx.stop()
		log.Printf("err: %v", err)
		return nil
	}
}

func mergeTokens(tokens []*token) *token {
	s := []string{}

	for _, tok := range tokens {
		s = append(s, tok.val)
	}

	return &token{
		tt:  tokenAtom,
		val: strings.Join(s, ""),
	}
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

func expectClosingBracket(p *parser) ([]*Node, error) {
	nodes := []*Node{}

loop:
	for {
		tok := p.next()

		switch tok.tt {
		case tokenHash:
			if _, err := expectComment(p); err != nil {
				return nil, err
			}
		case tokenOpenBracket:
			children, err := expectClosingBracket(p)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, &Node{
				Type:     NodeTypeArray,
				Children: children,
			})
		case tokenCloseBracket:
			break loop
		case tokenAtomSeparator, tokenNewLine:
			// ignore
		case tokenQuote:
			// collect
			tok, err := expectString(p)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, &Node{
				Type:  NodeTypeAtom,
				Value: tok.val,
			})
		case tokenAtom, tokenOpenList, tokenCloseList:
			nodes = append(nodes, &Node{
				Type:  NodeTypeAtom,
				Value: tok.val,
			})
		default:
			return nil, fmt.Errorf("unexpected token %v", tok)
		}
	}

	return nodes, nil
}

func expectDelimiter(p *parser) ([]*Node, error) {
	nodes := []*Node{}

loop:
	for {
		tok := p.next()

		switch tok.tt {
		case tokenOpenBracket:
			children, err := expectClosingBracket(p)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, &Node{
				Type:     NodeTypeArray,
				Children: children,
			})
		case tokenOpenList:
			children, err := expectDelimiter(p)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, &Node{
				Type:     NodeTypeList,
				Children: children,
			})
		case tokenCloseList:
			break loop
		case tokenAtomSeparator, tokenNewLine:
			// ignore
		case tokenQuote:
			// collect
			tok, err := expectString(p)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, &Node{
				Type:  NodeTypeAtom,
				Value: tok.val,
			})
		case tokenAtom:
			nodes = append(nodes, &Node{
				Type:  NodeTypeAtom,
				Value: tok.val,
			})
		default:
			return nil, fmt.Errorf("unexpected token %v", tok)
		}
	}

	return nodes, nil
}

func parserStateOpenBracket(p *parser) parserState {
	nodes, err := expectClosingBracket(p)
	if err != nil {
		return parserErrorState(err)
	}

	p.root.Children = append(p.root.Children, &Node{
		Type:     NodeTypeArray,
		Children: nodes,
	})
	return parserDefaultState
}

func parserStateOpenList(p *parser) parserState {
	nodes, err := expectDelimiter(p)
	if err != nil {
		return parserErrorState(err)
	}

	p.root.Children = append(p.root.Children, &Node{
		Type:     NodeTypeList,
		Children: nodes,
	})
	return parserDefaultState
}

func parse(in []byte) (*Node, error) {
	p := newParser(bytes.NewReader(in))

	err := p.run()
	if err != nil {
		return nil, err
	}

	printNodes(p.root.Children, 0)

	return p.root, nil
}

func printNodes(nodes []*Node, level int) {
	for _, node := range nodes {
		indent := strings.Repeat("\t", level)
		switch node.Type {
		case NodeTypeList, NodeTypeArray:
			fmt.Printf("%s(%s):\n", indent, nodeTypeName[node.Type])
			printNodes(node.Children, level+1)
		case NodeTypeAtom:
			fmt.Printf("%s(%s): %#v\n", indent, nodeTypeName[node.Type], node.Value)
		default:
			panic("unknown node type")
		}
	}
}
