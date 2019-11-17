package sexpr

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexer(t *testing.T) {
	tokens, err := tokenize([]byte(`(+ 1234 (- 566 6))`))
	log.Printf("tokens: %v", tokens)

	assert.NotNil(t, tokens)
	assert.NoError(t, err)
}
