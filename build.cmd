@echo off
setlocal enabledelayedexpansion

REM 删除dist
rmdir /s /q .\dist

REM 创建dist目录
mkdir .\dist

REM 定义平台数组
set platforms[0]=darwin/amd64
set platforms[1]=darwin/arm64
set platforms[2]=linux/amd64
set platforms[3]=linux/arm64
set platforms[4]=windows/amd64
set platforms[5]=windows/arm64


REM 版本号
set version=1.1.2

REM 遍历数组并执行操作
for /l %%i in (0,1,5) do (
    REM 分割平台字符串
    for /f "tokens=1,2 delims=/" %%a in ("!platforms[%%i]!") do (
        REM 构建平台
        set GOOS=%%a
        set GOARCH=%%b

        REM 构建可执行文件名
        set output_name=video-downloader-!GOOS!-!GOARCH!-%version%

        REM Windows平台特殊处理，添加.exe后缀
        if "!GOOS!"=="windows" set output_name=!output_name!.exe

        REM 编译
        go build -o .\dist\!output_name! main.go

        echo Built !output_name!
    )
)

echo Build process completed!
