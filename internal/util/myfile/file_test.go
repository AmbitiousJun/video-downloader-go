package myfile_test

import (
	"fmt"
	"log"
	"testing"
	"video-downloader-go/internal/util/myfile"
)

func TestDeleteDir(t *testing.T) {
	path := "c:/Users/Ambitious/Downloads/1.mp4_ts_dir"
	err := myfile.DeleteDir(path)
	if err != nil {
		t.Error("删除目录失败", err)
	}
}

func TestInitFileDir(t *testing.T) {
	path := "/Users/ambitious/Downloads/test/1.txt"
	err := myfile.InitFileDirs(path)
	if err != nil {
		t.Fatal(err)
	}
}

func TestInitTempTsDir(t *testing.T) {
	path := "c:/Users/Ambitious/Downloads/1.mp4"
	dirPath, err := myfile.InitTempTsDir(path, "temp_ts_dir")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(dirPath)
}

func TestDeleteAnyFileContainsName(t *testing.T) {
	p := "/Users/ambitious/Downloads/12138.mp4"
	s, d, err := myfile.DeleteAnyFileContainsPrefix(p)
	if err != nil {
		t.Error(err)
	}
	log.Printf("匹配个数：%d, 删除个数：%d", s, d)
}
