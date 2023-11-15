package file_test

import (
	"fmt"
	"testing"
	"video-downloader-go/src/util/file"
)

func TestDeleteDir(t *testing.T) {
	path := "c:/Users/Ambitious/Downloads/1.mp4_ts_dir"
	success := file.DeleteDir(path)
	if !success {
		t.Error("删除目录失败")
	}
}

func TestInitFileDir(t *testing.T) {
	path := "/Users/ambitious/Downloads/test/1.txt"
	err := file.InitFileDirs(path)
	if err != nil {
		t.Fatal(err)
	}
}

func TestInitTempTsDir(t *testing.T) {
	path := "c:/Users/Ambitious/Downloads/1.mp4"
	dirPath, err := file.InitTempTsDir(path)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(dirPath)
}
