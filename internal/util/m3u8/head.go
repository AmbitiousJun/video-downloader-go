// 处理特殊 m3u8 文件的 head 头
package m3u8

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"video-downloader-go/internal/util/mylog"
)

const (
	KeyTag = "m3u_key"
)

const (
	ExtXMap = "#EXT-X-MAP:"
)

// HeadInfo 存放从 m3u8 文件中解析出来的视频头部信息
type HeadInfo struct {
	Uri string `m3u_key:"URI" json:"URI"`
}

func (hi *HeadInfo) String() string {
	json, _ := json.Marshal(hi)
	return string(json)
}

// ResolveXMap 解析 m3u8 文件的 ExtXMap 头信息
// 接收 m3u8 文件的一行数据，如果解析成功，返回 HeadInfo 对象
// 解析失败则返回错误
func ResolveXMap(line string) (*HeadInfo, error) {
	// 1 检查前缀
	if !strings.HasPrefix(line, ExtXMap) {
		return nil, errors.New("不是正确的 EXT-X-MAP 格式: " + line)
	}
	line = strings.TrimPrefix(line, ExtXMap)

	// 2 逐字段解析
	headInfo := new(HeadInfo)
	kvs := strings.SplitN(line, `="`, -1)
	if len(kvs)%2 != 0 {
		return nil, errors.New("不是正确的 EXT-X-MAP 格式: " + line)
	}

	kvMap := make(map[string]string)
	for i := 0; i < len(kvs); i += 2 {
		key, value := kvs[i], kvs[i+1]
		kvMap[key] = strings.Trim(value, `"`)
	}

	v := reflect.ValueOf(headInfo).Elem()
	for i := 0; i < v.NumField(); i++ {
		// 获取当前属性的类型
		vType := v.Type().Field(i)

		// 获取属性的 key 标签
		t := vType.Tag.Get(KeyTag)
		if value, ok := kvMap[t]; ok {
			v.Field(i).SetString(value)
		}
	}

	mylog.Infof("EXT-X-MAP 解析结果: %v", headInfo)
	return headInfo, nil
}
