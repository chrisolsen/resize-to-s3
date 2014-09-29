package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"strconv"
)

type ImageTranscoder struct {
	path       string
	filename   string
	origHeight int
	origWidth  int
}

func NewImageTranscoder(id, filename string, file multipart.File) (*ImageTranscoder, error) {
	// create tmp path
	path, err := createTempDir(id)
	if err != nil {
		return nil, err
	}

	// save the original file
	output, err := os.Create(path + "/" + filename)
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
		path:       path,
		filename:   filename,
		origHeight: img.Bounds().Dy(),
		origWidth:  img.Bounds().Dx(),
	}, nil
}

// Cleanup deletes the folder containing all the image files
func (self *ImageTranscoder) Cleanup() error {
	return os.RemoveAll(self.path)
}

// ResizeTo resized the original image to each of the passed in sizes
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
	dims := fmt.Sprintf("%dx%d", size, size)
	savePath := self.path + "/" + strconv.Itoa(size) + "_" + self.filename
	srcPath := self.path + "/" + self.filename

	// TODO: add logic to resize based on image size
	cmd := exec.Command("convert",
		"-auto-orient",
		"-thumbnail", dims,
		"-gravity", "center",
		"-extent", dims,
		"-unsharp", "0x.9",
		srcPath,
		savePath,
	)

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return savePath, nil
}

// create a temp folder for the user to save temp files to
func createTempDir(id string) (string, error) {
	path := "/tmp/" + id
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", err
	}
	return path, nil
}
