package main

import (
	"errors"
	"io/ioutil"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

// TODO:
// * update the ACL, granting access to cloudfront
type S3Uploader struct {
	secretKey string
	accessKey string
	bucket    string
	region    string
}

// TODO: upload in channels
func (self *S3Uploader) Upload(rootPath string, filenames []string) error {
	for _, filename := range filenames {
		if err := self.upload(rootPath, filename); err != nil {
			return err
		}
	}
	return nil
}

func (self *S3Uploader) upload(rootPath, filename string) error {

	// read file
	data, err := ioutil.ReadFile(rootPath + "/" + filename)
	if err != nil {
		return err
	}

	// connect to s3
	auth := aws.Auth{
		AccessKey: self.accessKey,
		SecretKey: self.secretKey,
	}

	r, ok := aws.Regions[self.region]
	if !ok {
		return errors.New(self.region + " is an invalid region")
	}

	conn := s3.New(auth, r)
	b := conn.Bucket(self.bucket)

	// upload
	err = b.Put(filename, data, "image/png", s3.AuthenticatedRead)
	if err != nil {
		return err
	}

	return nil
}
