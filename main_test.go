package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func newUploadRequest(uri string, filePath string, formValues map[string]string) (*http.Request, error) {

	// where all data will be writted into
	body := &bytes.Buffer{}

	// obtain fileId
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// create multipart writer for full request
	writer := multipart.NewWriter(body)
	defer writer.Close()

	// obtain writer for file only
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}

	// write the file contents
	_, err = io.Copy(part, file)

	// write additional form params
	for key, val := range formValues {
		if err := writer.WriteField(key, val); err != nil {
			return nil, err
		}
	}

	// return request
	r, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}
	r.Header.Add("Content-Type", writer.FormDataContentType())

	return r, nil
}

// Make a full http request with a test image
func Test_FileSubmittal(t *testing.T) {

	params := map[string]string{
		"userId": "99",
		"sig":    "some_sig_value",
	}

	path, _ := os.Getwd()
	path += "/files/test.jpg"
	req, err := newUploadRequest("/attach", path, params)
	res := httptest.NewRecorder()

	http.HandlerFunc(AttachImage).ServeHTTP(res, req)
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err.Error())
	}

	fmt.Println("JSON:", string(data))

	// expect a json object containg the resize image data
	var details resizeDetails
	if err := json.Unmarshal(data, &details); err != nil {
		t.Fatal(err.Error())
	}
	// validate expected data was returned
	if details.ThumbnailSmallName != "test_small.jpg" {
		t.Error("Invalid small file name")
	}

	if details.ThumbnailMediumName != "test_medium.jpg" {
		t.Error("Invalid medium file name")
	}

	if details.ThumbnailLargeName != "test_large.jpg" {
		t.Error("Invalid large file name")
	}

	if details.ThumbnailXLargeName != "test_xlarge.jpg" {
		t.Error("Invalid extra large file name")
	}

	if details.StandardSize != 493654 {
		t.Error("Invalid size")
	}

	if details.ContentType != "image/jpg" {
		t.Error("Invalid content type")
	}

}
