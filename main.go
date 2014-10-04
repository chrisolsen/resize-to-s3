package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// api response structure (uses pascal case to stay consistent with current api)
type resizeDetails struct {
	ThumbnailSmallName  string
	ThumbnailMediumName string
	ThumbnailLargeName  string
	ThumbnailXLargeName string
	StandardName        string
	StandardSize        int
	ContentType         string
}

// ./config/settings.json struct
// TODO: move, at the very least, the secret key into an ENV var
type config struct {
	Sizes []int `json:"sizes"`
	S3    struct {
		SecretKey string `json:"secret_key"`
		AccessKey string `json:"access_key"`
		Bucket    string `json:"bucket"`
		Region    string `json:"region"`
	} `json:"s3"`
}

var settings config

func init() {
	b, err := ioutil.ReadFile("./config/settings.json")
	if err != nil {
		panic("Unable to load ./config/settings.json file")
	}

	json.NewDecoder(bytes.NewReader(b)).Decode(&settings)
}

// Server
func main() {
	r := http.NewServeMux()

	r.HandleFunc("/attach", AttachImage)
	http.ListenAndServe(":3000", r)
}

// AttachImage resizes the attached image to the set sizes and uploads them
// to the S3 bucket specified in the settings.json file.
// AttachImage expects a multipart/form-data request to be made with the
// following parameters:
// 	* userId
//  * sig
// 	* file
func AttachImage(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Must be a POST request", http.StatusMethodNotAllowed)
		return
	}

	// initialization
	userId := r.FormValue("userId")
	// sig := r.FormValue("sig")
	s3 := &S3Uploader{
		secretKey: settings.S3.SecretKey,
		accessKey: settings.S3.AccessKey,
		bucket:    settings.S3.Bucket,
		region:    settings.S3.Region,
	}

	// obtain file details
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, toErr("Bad file", err), http.StatusInternalServerError)
		return
	}

	// resize
	t, err := NewImageTranscoder("/tmp", userId, file, fileHeader)
	if err != nil {
		http.Error(w, toErr("Transcoder Init Fail", err), http.StatusInternalServerError)
		return
	}

	filenames, err := t.ResizeTo(30, 60, 80, 160)
	if err != nil {
		http.Error(w, toErr("Resize Fail", err), http.StatusInternalServerError)
		return
	}

	// upload files
	if err := s3.Upload("/tmp", filenames); err != nil {
		http.Error(w, toErr("S3 Fail", err), http.StatusInternalServerError)
		return
	}

	// clean up dir and files
	if err := t.Cleanup(); err != nil {
		http.Error(w, toErr("Removing dir", err), http.StatusInternalServerError)
		return
	}

	// send response
	details := resizeDetails{
		ThumbnailSmallName:  "test_small.jpg",
		ThumbnailMediumName: "test_medium.jpg",
		ThumbnailLargeName:  "test_large.jpg",
		ThumbnailXLargeName: "test_xlarge.jpg",
		StandardName:        "test.jpg",
		StandardSize:        493654,
		ContentType:         "image/jpg",
	}

	if err := json.NewEncoder(w).Encode(details); err != nil {
		http.Error(w, toErr("Creating json response", err), http.StatusInternalServerError)
		return
	}
}

// Helper to package err messages in a json struct
func toErr(msg string, err error) string {
	errMsg := "-"
	if err != nil {
		errMsg = err.Error()
	}

	b, _ := json.Marshal(map[string]string{
		"msg":   msg,
		"error": errMsg,
	})
	return string(b)
}
