package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime/multipart"
	"os"
	"os/exec"
	"strconv"
)

type ImageTranscoder struct {
	tempPath   string
	id         string
	filename   string
	origHeight int
	origWidth  int
	mimetype   string
}

func NewImageTranscoder(tmpPath, id string, file multipart.File, fileHeader *multipart.FileHeader) (*ImageTranscoder, error) {

	// create tmp path
	tmp, err := createTempDir(tmpPath, id)
	if err != nil {
		return nil, err
	}

	// save the original file
	output, err := os.Create(tmp + "/" + fileHeader.Filename)
	if err != nil {
		return nil, err
	}
	defer output.Close()
	if _, err := io.Copy(output, file); err != nil {
		return nil, err
	}

	// obtain dims of original file
	file.Seek(0, 0)
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return &ImageTranscoder{
		id:         id,
		tempPath:   tmpPath,
		filename:   fileHeader.Filename,
		origHeight: img.Bounds().Dy(),
		origWidth:  img.Bounds().Dx(),
		mimetype:   "image/png",
	}, nil
}

// Cleanup deletes the folder containing all the image files
func (self *ImageTranscoder) Cleanup() error {
	return os.RemoveAll(self.tempPath + "/" + self.id)
}

// ResizeTo resized the original image to each of the passed in sizes
// TODO: resize in channels
func (self *ImageTranscoder) ResizeTo(sizes ...int) ([]string, error) {
	filenames := make([]string, len(sizes))
	for i, size := range sizes {
		filename, err := self.resizeTo(size)
		if err != nil {
			return nil, err
		}
		filenames[i] = filename
	}

	return filenames, nil
}

// Resize the image to the specified size.
func (self *ImageTranscoder) resizeTo(size int) (string, error) {
	log.Println("resizing to", strconv.Itoa(size))
	dims := fmt.Sprintf("%dx%d", size, size)
	sizedFilename := self.id + "/" + strconv.Itoa(size) + "_" + self.filename

	// TODO: add logic to resize based on image size
	cmd := exec.Command("convert",
		"-auto-orient",
		"-thumbnail", dims,
		"-gravity", "center",
		"-extent", dims,
		"-unsharp", "0x.9",
		self.tempPath+"/"+self.id+"/"+self.filename,
		self.tempPath+"/"+sizedFilename,
	)

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return sizedFilename, nil
}

// create a temp folder for the user to save temp files to
func createTempDir(tmp, id string) (string, error) {
	path := tmp + "/" + id
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", err
	}
	return path, nil
}
