package sexpr

import (
	"bytes"
	"io"
	"log"
)

type Reader struct {
	r io.Reader
}

func Parse(in []byte) (*Node, error) {
	r := NewReader(bytes.NewReader(in))
	return r.Parse()
}

func NewReader(r io.Reader) *Reader {
	return &Reader{r: r}
}

func (r *Reader) Parse() (*Node, error) {
	lx := newLexer(r.r)
	go func() {
		for tok := range lx.tokens {
			log.Printf("tok: %v -- %#v", tok, tok)
		}
	}()
	err := lx.run()
	return nil, err
}
