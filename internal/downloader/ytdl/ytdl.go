package ytdl

import (
	"fmt"
	"os/exec"
	"strings"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/downloader/coredl"
	"video-downloader-go/internal/meta"
	"video-downloader-go/internal/util/m3u8"
	"video-downloader-go/internal/util/myfile"
	"video-downloader-go/internal/util/mylog"

	"github.com/pkg/errors"
)

// YtDlDownloader 是适配 youtube-dl 的下载器
// 实现了 coredl.Downloader 接口
type YtDlDownloader struct {
	mp4Dl  coredl.Downloader
	m3u8Dl coredl.Downloader
}

// New 初始化一个 youtube-dl 下载器
func New() coredl.Downloader {
	dl := new(YtDlDownloader)
	if config.G.Downloader.Use == config.DownloadMultiThread {
		dl.m3u8Dl = coredl.NewM3U8MultiThread()
		dl.mp4Dl = coredl.NewMp4MultiThread()
		return dl
	}
	dl.m3u8Dl = coredl.NewM3U8Simple()
	dl.mp4Dl = coredl.NewMp4Simple()
	return dl
}

// Exec 是实现 coredl.Downloader 的核心下载逻辑
func (d *YtDlDownloader) Exec(dmt *meta.Download, handlerFunc coredl.ProgressHandler) error {
	// 1 恢复下载信息
	links := meta.Split2YtDlLinks(dmt.Link)
	size := len(links)

	// 2 拆分任务
	progressHandler := func(curTask int) coredl.ProgressHandler {
		return func(p *coredl.Progress) {
			p.CurrentTask, p.TotalTasks = curTask, size
			handlerFunc(p)
		}
	}

	for i, link := range links {
		mylog.Infof("正在处理第 %d / %d 个子任务，文件名：%s", i+1, size, dmt.FileName)
		tmpDmt := meta.NewDownloadMeta(link, strings.Replace(dmt.FileName, ".mp4", getFilePartSuffix(i), -1), dmt.OriginUrl)
		tmpDmt.LogBar = dmt.LogBar

		var err error
		if m3u8.CheckM3U8(link, dmt.HeaderMap) {
			err = d.m3u8Dl.Exec(tmpDmt, progressHandler(i+1))
		} else {
			err = d.mp4Dl.Exec(tmpDmt, progressHandler(i+1))
		}

		if err != nil {
			return err
		}

		mylog.Successf("第 %d / %d 个子任务处理完成，文件名：%s", i+1, size, dmt.FileName)
	}

	if err := mergeSubTask(dmt, size); err != nil {
		return errors.Wrap(err, "合并子任务失败")
	}

	return nil
}

// mergeSubTask 调用 ffmpeg 将子任务合并在一起
func mergeSubTask(dmt *meta.Download, size int) error {
	mylog.Infof("正在合并子任务，文件名：%s", dmt.FileName)

	// 输入需要合并的子任务
	commands := []string{}
	for i := 0; i < size; i++ {
		commands = append(commands, "-i", strings.Replace(dmt.FileName, ".mp4", getFilePartSuffix(i), -1))
	}

	// 直接拷贝视频流和音频流，不进行转码
	commands = append(commands, "-c:v", "copy", "-c:a", "copy", dmt.FileName)

	// 执行命令
	cmd := exec.Command(config.FfmpegPath, commands...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "合并命令执行失败")
	}

	invalidInput := "Invalid data found when processing input"
	if strings.Contains(string(output), invalidInput) {
		mylog.Errorf("合并子任务失败，子任务不自动删除，文件名：%s", dmt.FileName)
	}

	mylog.Successf("子任务合并完成，正在删除子任务，文件名：%s", dmt.FileName)
	flag := true
	for i := 0; i < size; i++ {
		partName := strings.Replace(dmt.FileName, ".mp4", getFilePartSuffix(i), 1)
		if e, d := myfile.DeleteFileIfExist(partName); e && !d {
			mylog.Warnf("子任务删除失败，文件名：%s", partName)
			flag = false
		}
	}
	if flag {
		mylog.Success("所有子任务已全部删除")
	}
	return nil
}

// getFilePartSuffix 根据子任务索引返回子任务的文件后缀
func getFilePartSuffix(i int) string {
	return fmt.Sprintf("_part%d.mp4", i)
}
