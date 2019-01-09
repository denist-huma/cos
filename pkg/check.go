package pkg

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/dmolesUC3/cos/internal"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// The Check struct represents a fixity check operation
type Check struct {
	Logger    internal.Logger
	ObjLoc    internal.ObjectLocation
	Expected  []byte
	Algorithm string
	Region    string
}

// GetDigest gets the digest, returning an error if the object cannot be retrieved or,
// when an expected digest is provided, if the calculated digest does not match.
func (c Check) GetDigest() ([]byte, error) {
	logger := c.Logger
	objLoc := c.ObjLoc

	logger.Detail("Initializing session")
	sess, err := objLoc.GetSession()
	if err != nil {
		return nil, err
	}

	s3Object, err := objLoc.GetS3Object()
	if err != nil {
		return nil, err
	}
	logger.Detailf("Expected ContentLength: %d\n", *s3Object.ContentLength)

	// TODO: don't write to tempfile
	filename := path.Base(*objLoc.Key())
	outfile, err := ioutil.TempFile("", filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := os.Remove(outfile.Name())
		if err != nil {
			logger.Info(err)
		}
	}()
	logger.Detailf("Downloading to tempfile: %v\n", outfile.Name())
	downloader := s3manager.NewDownloader(sess)
	bytesDownloaded, err := downloader.Download(outfile, &s3.GetObjectInput{
		Bucket: objLoc.Bucket(),
		Key:    objLoc.Key(),
	})
	logger.Detailf("Downloaded %d bytes\n", bytesDownloaded)
	if err != nil {
		return nil, err
	}
	err = outfile.Close() // TODO is this necessary?
	if err != nil {
		return nil, err
	}

	infile, err := os.Open(outfile.Name())
	if err != nil {
		return nil, err
	}
	h := c.newHash()
	bytesHashed, err := io.Copy(h, infile)
	logger.Detailf("Hashed %d bytes\n", bytesHashed)
	if err != nil {
		return nil, err
	}
	digest := h.Sum(nil)

	if len(c.Expected) > 0 {
		if !bytes.Equal(c.Expected, digest) {
			err = fmt.Errorf("digest mismatch: expected: %x, actual: %x", c.Expected, digest)
		}
	}

	return digest, err
}

func (c Check) newHash() hash.Hash {
	if c.Algorithm == "sha256" {
		return sha256.New()
	}
	return md5.New()
}



