package cos

import (
	"crypto/md5"
	"crypto/sha256"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const DefaultAwsRegion = "us-west-2"
var awsRegionRegexp = regexp.MustCompile("https?://s3-([^.]+)\\.amazonaws\\.com/")

type Fixity struct {
	Logger    Logger
	ObjLoc    ObjectLocation
	Expected  []byte
	Algorithm string
	Region    string
}


func (f Fixity) GetDigest() ([]byte, error) {
	sess, err := f.initSession()
	if err != nil {
		return nil, err
	}

	// TODO: don't write to tempfile
	outfile, err := ioutil.TempFile("", f.objFilename())
	if err != nil {
		return nil, err
	}
	// TODO: uncomment this
	//defer func() {
	//	err := os.Remove(outfile.Name())
	//	if err != nil {
	//		f.Logger.Info(err)
	//	}
	//}()
	f.Logger.Detailf("Downloading to tempfile: %v\n", outfile.Name())

	downloader := s3manager.NewDownloader(sess)
	bytesDownloaded, err := downloader.Download(outfile, &s3.GetObjectInput{
		Bucket: f.bucketP(),
		Key:    f.keyP(),
	})
	f.Logger.Detailf("Downloaded %d bytes\n", bytesDownloaded)
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

	h := f.newHash()
	bytesHashed, err := io.Copy(h, infile)
	f.Logger.Detailf("Hashed %d bytes\n", bytesHashed)
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func (f Fixity) newHash() hash.Hash {
	if f.Algorithm == "sha256" {
		return sha256.New()
	}
	return md5.New()
}

func (f Fixity) initSession() (*session.Session, error) {
	f.Logger.Detail("Initializing session")
	endpointP := f.endpointP()
	s3Config := aws.Config{
		Endpoint: endpointP,
		Region: f.regionStrP(),
		S3ForcePathStyle: aws.Bool(true),
	}
	s3Opts := session.Options{
		Config:            s3Config,
		SharedConfigState: session.SharedConfigEnable,
	}
	return session.NewSessionWithOptions(s3Opts)
}

func (f Fixity) regionStrP() *string {
	if f.Region != "" {
		f.Logger.Detailf("Using specified AWS region: %v\n", f.Region)
		return &f.Region
	}
	endpoint := f.endpointStr()
	matches := awsRegionRegexp.FindStringSubmatch(endpoint)
	regionStr := DefaultAwsRegion
	if len(matches) == 2 {
		f.Logger.Detailf("Found AWS region in endpoint URL %v: %v\n", endpoint, regionStr)
		regionStr = matches[1]
	} else {
		f.Logger.Detailf("No AWS region found in endpoint URL %v; using default region %v\n", endpoint, DefaultAwsRegion)
	}
	return &regionStr
}

func (f Fixity) endpointStr() string {
	return f.ObjLoc.Endpoint.String()
}

func (f Fixity) endpointP() *string {
	endpointUrlStr := f.endpointStr()
	return &endpointUrlStr
}

func (f Fixity) objFilename() string {
	return path.Base(f.ObjLoc.Key())
}

func (f Fixity) bucketP() *string {
	bucket := f.ObjLoc.Bucket()
	return &bucket
}

func (f Fixity) keyP() *string {
	key := f.ObjLoc.Key()
	return &key
}
