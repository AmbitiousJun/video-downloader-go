# video-downloader-go

这是 [video-downloader](https://github.com/AmbitiousJun/video-downloader) 的 Go 实现版本



### 特点

1. 打包之后体积更小，无需 JVM 环境也能使用

2. 功能基本上都迁移过来了，解析器只实现了 youtube-dl，与 Selenium 有关的解析器还没有实现

3. 美化终端输出



### 概述

使用 Go 语言编写的多线程视频下载器，适配 “爱优腾芒”。开发这个项目的目的就是为了**批量下载**视频的时候解放双手，不需要手动转换 m3u8，也不需要等到视频下载完成之后再去一个一个改名。

一句话总结这个项目：类似 docker-compose，本项目就是将下载的任务以及下载方式提前通过配置的方式编排好，然后启动程序自动下载。



### 适用场景

- 批量下载视频
- 文件名称提前配置
- 自动将 ts 文件合并成 mp4
- **需要给视频文件标准命名以生成海报墙**（Emby, Jellyfin, Infuse, Kodi）

![架构图](./img/3.jpg)

## 技术栈

- Go



## 安装&使用

1. 下载适配自己系统的压缩包，解压后存放到自定义位置即可

2. 修改 config.yml

默认情况下，转换器保持 ffmpeg 的配置不需要改变。

修改解析器和下载器的配置即可

3. 修改 data.txt

在这个文件中编写下载任务，每一行是一个任务，格式：`文件名｜url`，文件名不需要包含扩展名，下载默认为 `mp4`。

4. 启动程序

打开终端，定位到 video-downloader-go 根目录，执行：

> 在 macos / linux 环境下，可能会报错 **ffmpeg, youtube-dl 检测失败**，这是因为可执行文件没有授予可执行权限。
> 
> 以 ffmpeg 为例，定位到文件路径，并分配可执行权限即可：
> 
> ```shell
> cd ./config/ffmpeg
> chmod +x ./ffmpeg-macos
> ```

```shell
# macos / linux
./start

# windows
start.exe
```

## 示例

1. 不使用解析器，多线程下载 mp4 格式视频

data.txt:

```shell
这是一个视频|https://example.com/test.mp4
```

config.yml:

```yml
decoder: # 解码器相关配置
  use: none # 使用哪种解析方式，可选值：none, free-api, vip-fetch, youtube-dl，若使用 youtube-dl，resource-type 会被忽略
  resource-type: mp4 # 解析出来的文件类型，可选值：mp4, m3u8

downloader:
  use: multi-thread # 要使用哪个下载器，可选值：simple, multi-thread
  task-thread-count: 1 # 处理下载任务的线程个数
  dl-thread-count: 32 # 多线程下载的线程个数
  download-dir: /Users/ambitious/Downloads # 视频文件下载位置
  ts-dir-suffix: temp_ts_files # 暂存 ts 文件的目录后缀【保持默认即可】
```

2. 不使用解析器，多线程下载 m3u8 视频，并自动合并为 mp4

data.txt:

```shell
这是一个视频|https://example.com/test.m3u8
```

config.yml:

```yml
decoder: # 解码器相关配置
  use: none # 使用哪种解析方式，可选值：none, free-api, vip-fetch, youtube-dl，若使用 youtube-dl，resource-type 会被忽略
  resource-type: m3u8 # 解析出来的文件类型，可选值：mp4, m3u8

downloader:
  use: multi-thread # 要使用哪个下载器，可选值：simple, multi-thread
  task-thread-count: 1 # 处理下载任务的线程个数
  dl-thread-count: 32 # 多线程下载的线程个数
  download-dir: /Users/ambitious/Downloads # 视频文件下载位置
  ts-dir-suffix: temp_ts_files # 暂存 ts 文件的目录后缀

transfer:
  use: ffmpeg # 要选用哪个转码器，可选值：file-channel, cv, ffmpeg【保持ffmpeg不变即可】
  ts-filename-regex: (?<=_)(\d+)(?=\.) # 正则表达式，用于匹配出 ts 文件的序号
```

3. 已有 “爱优腾芒” 等视频网站的会员，需要批量下载网站上的视频

data.txt:

```shell
开始推理吧.S01E01|https://v.qq.com/x/cover/mzc00200ynivua7/r00434mq14v.html
开始推理吧.S01E02|https://v.qq.com/x/cover/mzc00200ynivua7/r00434mq14v.html
开始推理吧.S01E03|https://v.qq.com/x/cover/mzc00200ynivua7/r00434mq14v.html
开始推理吧.S01E04|https://v.qq.com/x/cover/mzc00200ynivua7/r00434mq14v.html
开始推理吧.S01E05|https://v.qq.com/x/cover/mzc00200ynivua7/r00434mq14v.html
开始推理吧.S01E06|https://v.qq.com/x/cover/mzc00200ynivua7/r00434mq14v.html
开始推理吧.S01E07|https://v.qq.com/x/cover/mzc00200ynivua7/r00434mq14v.html
开始推理吧.S01E08|https://v.qq.com/x/cover/mzc00200ynivua7/r00434mq14v.html
开始推理吧.S01E09|https://v.qq.com/x/cover/mzc00200ynivua7/r00434mq14v.html
开始推理吧.S01E10|https://v.qq.com/x/cover/mzc00200ynivua7/r00434mq14v.html
```

仍然是以 TX 为例，首先选取要下载的视频格式，在终端上运行：

```shell
youtube-dl -F "https://v.qq.com/x/cover/mzc00200ynivua7/r00434mq14v.html" --cookies-from-browser chrome
```

如果是会员才能观看的视频，需要先在浏览器登录会员账号，并注入 cookie，我这里以 chrome 为例，运行结果：

![分析youtube-dl的code](./img/4.jpg)

我想优先下载 1080p 格式，如果该格式下载失败的话，就下载 720p 的，那么配置文件这么写：

config.yml:

```yml
decoder: # 解码器相关配置
  use: youtube-dl # 使用哪种解析方式，可选值：none, free-api, vip-fetch, youtube-dl，若使用 youtube-dl，resource-type 会被忽略
  youtube-dl: # youtube-dl 解析器相关配置
    cookies-from: chrome # 从哪个浏览器获取 cookie，该参数会直接传递给 youtube-dl，传入 none 则忽略
    format-codes: # 下载视频的编码，可传多个，按照顺序进行解析，两种格式：'视频编码+音频编码' 或者 '视频编码'，只会下载首次解析成功的格式
      - fhd-0
      - shd-1

downloader:
  use: multi-thread # 要使用哪个下载器，可选值：simple, multi-thread
  task-thread-count: 1 # 处理下载任务的线程个数
  dl-thread-count: 32 # 多线程下载的线程个数
  download-dir: /Users/ambitious/Downloads # 视频文件下载位置
  ts-dir-suffix: temp_ts_files # 暂存 ts 文件的目录后缀

transfer:
  use: ffmpeg # 要选用哪个转码器，可选值：file-channel, cv, ffmpeg
  ts-filename-regex: (?<=_)(\d+)(?=\.) # 正则表达式，用于匹配出 ts 文件的序号
```

6. 已有 “爱优腾芒” 等视频网站的会员，需要批量下载网站上的视频，但是要下载的视频太多，懒得自己一个一个获取 format code

大多数视频网站中，通常情况下相同系列的视频相同格式它的 format code 是一样的，只需提前配置好一个 format code，就能解析下载全部视频。

但是像 **MG** 就不行了，每个视频的 format code 都是随机的，要下载 40 个视频，就要手动获取 40 个 format code，**非常地不银杏**。

这个时候就可以用到程序的自动获取 format code 功能了，当 config.yml 中配置的 format code 全部解析失败时，会触发这个逻辑：

![程序自动获取 format code](./img/5.jpg)

如果不想要自己提前手动获取 format code，那么 config.yml 中，`downloader.youtube-dl.format-codes` 配置就不需要传递任何内容，像这样：

```yml
decoder: # 解码器相关配置
  use: youtube-dl # 使用哪种解析方式，可选值：none, free-api, vip-fetch, youtube-dl，若使用 youtube-dl，resource-type 会被忽略
  youtube-dl: # youtube-dl 解析器相关配置
    cookies-from: chrome # 从哪个浏览器获取 cookie，该参数会直接传递给 youtube-dl，传入 none 则忽略
    format-codes: # 下载视频的编码，可传多个，按照顺序进行解析，两种格式：'视频编码+音频编码' 或者 '视频编码'，只会下载首次解析成功的格式
```

有的时候会因为网络问题导致 format code 生成异常，可以直接敲回车重新获取。



**记住已选择的视频格式：**

批量下载 MG 上的视频时，尽管程序已经提供了自动读取 format code 功能，但是当下载量较大时，还是需要人为频繁地手动输入 format code。

这时可以将 `decoder.youtube-dl.remember-format` 配置设置成 `1`，开启记住已选择的视频格式。



```yml
decoder:
  use: none # 使用哪种解析方式，可选值：none, youtube-dl，若使用 youtube-dl，resource-type 会被忽略
  resource-type: m3u8 # 解析出来的文件类型，可选值：mp4, m3u8
  youtube-dl: # youtube-dl 解析器相关配置
    cookies-from: firefox # 从哪个浏览器获取 cookie，推荐 firefox，该参数会直接传递给 youtube-dl，传入 none 则忽略
    format-codes: # 下载视频的编码，可传多个，按照顺序进行解析，两种格式：'视频编码+音频编码' 或者 '视频编码'，只会下载首次解析成功的格式，可以不传此参数，在程序执行时手动选择
    remember-format: 1 # 是否记住视频格式，程序自动根据 host 进行区分，每次启动程序时缓存都会重置，可选值：-1, 1
```



程序会在用户第一次输入 format code 的时候，记住该视频格式（自动根据 url host 进行区分），

在之后读取 format code 的时候，程序会自动进行匹配，匹配成功则自动进行解析，若失败，则依旧是手动输入。

> 有的网站使用 youtube-dl 解析出来的视频格式中，不同的 format code 的格式是一样的，程序会按照从上到下按顺序匹配，并使用最先匹配到的结果。



7. 对不同的网站进行定制化配置

如果想要不同的网站下载任务同时开始进行，而不同网站使用的解析器又不相同，或者不完全相同时，可以采用定制化配置，通过 `host` 来区分配置。

可以在 `customs`  属性中配置多个定制化配置，在 `customs.hosts` 属性下配置要匹配的域名，参考配置如下：



```yml
# 针对不同的域名进行定制化配置
# 
# 目前只支持针对 decoder 进行定制化配置
# 可配置的属性：use, resource-type, youtube-dl.cookies-from, youtube-dl.format-codes, youtube-dl.remember-format
customs:
  - decoder: 
      use: youtube-dl
      youtube-dl:
        cookies-from: firefox
        format-codes:
        remember-format: 1
    hosts: # 对哪些域名生效，必须配置完整，有端口也要加上
      - www.mgtv.com
      - www.youtube.com
      - www.bilibili.com
```

> 注：目前仅支持对解析器进行定制化配置