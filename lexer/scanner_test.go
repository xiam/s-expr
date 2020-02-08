package lexer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScannerStop(t *testing.T) {
	lx := New(bytes.NewReader([]byte(`1 2 3 4 5`)))

	errCh := make(chan error)
	go func() {
		errCh <- lx.Scan()
	}()

	go func() {
		for lx.Next() {
			_ = lx.Token()

			lx.Stop()
		}
	}()

	err := <-errCh
	assert.Equal(t, ErrForceStopped, err)
}
