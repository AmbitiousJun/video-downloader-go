package file_test

import (
	"testing"
	"video-downloader-go/src/util/file"
)

func TestInitFileDirs(t *testing.T) {
	path := "/Users/ambitious/Downloads/test/1.txt"
	err := file.InitFileDirs(path)
	if err != nil {
		t.Fatal(err)
	}
}
