package ast

import (
	"testing"
)

func TestPrintEmptyNode(t *testing.T) {
	empty := &Node{}
	Print(empty)
}
