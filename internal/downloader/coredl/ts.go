// ts 文件下载相关
package coredl

import (
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"time"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/util"
	"video-downloader-go/internal/util/m3u8"
	"video-downloader-go/internal/util/myfile"
	"video-downloader-go/internal/util/myhttp"
	"video-downloader-go/internal/util/mylog"

	"github.com/pkg/errors"
)

// Ts 文件处理器
type TsHandler struct {
	m3u8.TsMeta                   // ts 文件下载信息
	DlPath      string            // ts 文件保存的绝对路径
	Headers     map[string]string // 请求头

	valid       bool   // 当前处理器是否有效
	headless    bool   // 是否是无头下载
	dlDir       string // ts 文件下载目录
	tmpHeadName string // 暂存头部的文件名
	tmpBodyName string // 暂存主体的文件名

	// 重试机制，防止下载过快导致 ffmpeg 合并失败
	currentTry int // 当前尝试次数
	maxRetry   int // 最多尝试次数
}

// NewTsHandler 创建一个 ts 文件处理器
func NewTsHandler(tmt *m3u8.TsMeta, dlPath string, headers map[string]string) *TsHandler {
	th := &TsHandler{
		TsMeta:     *tmt,
		DlPath:     dlPath,
		Headers:    headers,
		valid:      true,
		currentTry: 1,
		maxRetry:   5,
	}

	th.dlDir = filepath.Dir(dlPath)
	th.headless = th.TsMeta.HeadUrl == ""
	if !th.headless {
		// 初始化头部和主体的暂存文件名
		th.tmpHeadName = util.RandString(32) + ".ts"
		th.tmpBodyName = util.RandString(32) + ".ts"
	}

	return th
}

// Download 执行下载逻辑
func (th *TsHandler) Download() (int64, error) {
	if !th.valid {
		return -1, errors.New("当前处理器已失效，请重新创建处理器进行下载")
	}

	defer func() {
		th.valid = false
	}()

	if th.headless {
		return th.downloadHeadless()
	}
	return th.downloadAndMergeHead()
}

// downloadAndMergeHead 下载 ts 头部和主体，并进行合并
func (th *TsHandler) downloadAndMergeHead() (int64, error) {
	dlDir := filepath.Dir(th.DlPath)
	defer func() {
		// 删除临时头部和主体
		if e, d := myfile.DeleteFileIfExist(filepath.Join(th.dlDir, th.tmpHeadName)); e && !d {
			mylog.Warnf("临时文件删除失败: %s", th.tmpHeadName)
		}
		if e, d := myfile.DeleteFileIfExist(filepath.Join(th.dlDir, th.tmpBodyName)); e && !d {
			mylog.Warnf("临时文件删除失败: %s", th.tmpBodyName)
		}
	}()

	// 1 下载头部保存为一个临时 ts
	req, err := th.buildRequestWithHeaders(true)
	if err != nil {
		return -1, err
	}
	headDn, err := myhttp.DownloadWithRateLimitV2(req, filepath.Join(dlDir, th.tmpHeadName))
	if err != nil {
		return -1, errors.Wrapf(err, "分片下载异常: %v", th.DlPath)
	}

	// 2 下载主体保存为另一个临时 ts
	req, err = th.buildRequestWithHeaders(false)
	if err != nil {
		return -1, err
	}
	bodyDn, err := myhttp.DownloadWithRateLimitV2(req, filepath.Join(dlDir, th.tmpBodyName))
	if err != nil {
		return -1, errors.Wrapf(err, "分片下载异常: %v", th.DlPath)
	}

	// 3 将头部和主体使用 ffmpeg 进行合并到 dlPath
	if err = th.mergeHeadAndBody(); err != nil {
		if th.currentTry < th.maxRetry {
			time.Sleep(time.Second * 2)
			th.currentTry++
			return th.downloadAndMergeHead()
		}
		return -1, err
	}

	return headDn + bodyDn, nil
}

// downloadHeadless 下载无头的 ts 分片
func (th *TsHandler) downloadHeadless() (int64, error) {
	req, err := th.buildRequestWithHeaders(false)
	if err != nil {
		return -1, err
	}

	var dn int64
	if dn, err = myhttp.DownloadWithRateLimitV2(req, th.DlPath); err != nil {
		return -1, errors.Wrapf(err, "分片下载异常：%v", th.DlPath)
	}

	return dn, nil
}

// buildRequestWithHeaders 构造一个携带请求头的 http 请求对象
// 接收一个布尔参数，为 true 时下载 ts 头部，否则下载 ts 主体
func (th *TsHandler) buildRequestWithHeaders(head bool) (*http.Request, error) {
	url := th.TsMeta.Url
	if head {
		url = th.TsMeta.HeadUrl
	}

	var req *http.Request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "构造请求时出现异常：%v", th.DlPath)
	}

	for k, v := range th.Headers {
		req.Header.Add(k, v)
	}

	return req, nil
}

// mergeHeadAndBody 使用 ffmpeg 将 ts 头部和主体合并到一起
func (th *TsHandler) mergeHeadAndBody() error {
	// 1 构建命令
	cmd := exec.Command(
		config.FfmpegPath,
		"-i", fmt.Sprintf("concat:%s|%s", filepath.Join(th.dlDir, th.tmpHeadName), filepath.Join(th.dlDir, th.tmpBodyName)),
		"-c", "copy",
		th.DlPath,
	)

	// 2 执行合并
	if _, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "分片下载异常: %v", th.DlPath)
	}

	return nil
}
