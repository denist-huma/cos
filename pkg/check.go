package pkg

import (
	"bytes"
	"fmt"

	"github.com/dmolesUC3/cos/internal/objects"
	"github.com/dmolesUC3/cos/internal/streaming"
)

// The Check struct represents a fixity check operation
type Check struct {
	Object    objects.Object
	Expected  []byte
	Algorithm string
}

// CalcDigest gets the digest, returning an error if the object cannot be retrieved or,
// when an expected digest is provided, if the calculated digest does not match.
func (c Check) CalcDigest() ([]byte, error) {
	actualDigest, err := objects.CalcDigest(c.Object, streaming.DefaultRangeSize, c.Algorithm)
	if err != nil {
		return nil, err
	}
	expectedDigest := c.Expected
	if len(expectedDigest) > 0 {
		if !bytes.Equal(expectedDigest, actualDigest) {
			err = fmt.Errorf("digest mismatch: expected: %x, actual: %x", expectedDigest, actualDigest)
		}
	}
	return actualDigest, err
}
