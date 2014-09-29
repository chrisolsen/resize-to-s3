package main

// TODO:
// * upload the file with putObject
// * update the ACL, granting access to cloudfront
type S3Uploader struct {
	secretKey string
	accessKey string
	bucket    string
}

func (self *S3Uploader) Upload(filenames []string) error {
	for _, filename := range filenames {
		self.upload(filename)
	}
	return nil
}

func (self *S3Uploader) upload(filename string) error {
	// s := s3.New(auth, region)
	// bucket := s.Bucket("bucket_name")
	// bucket.Put("key", bytes, contentType, perm)

	return nil
}
