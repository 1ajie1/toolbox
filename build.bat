@echo off
setlocal enabledelayedexpansion

REM 显示帮助信息的函数
:showHelp
    echo 运维工具箱跨平台编译脚本
    echo.
    echo 用法: %0 [选项]
    echo 选项:
    echo   /h, /help            显示帮助信息
    echo   /o, /os OS_NAME      指定目标操作系统 (windows, linux, darwin)
    echo   /a, /all             编译所有支持的操作系统
    echo   /v, /version VERSION 指定版本号 (默认: 1.0.0)
    echo.
    echo 示例:
    echo   %0 /os windows       仅编译Windows版本
    echo   %0 /all              编译所有操作系统的版本
    echo   %0 /all /version 1.2.3   编译所有操作系统的1.2.3版本
    exit /b 0

REM 默认值
set "VERSION=1.0.0"
set "BUILD_ALL=false"
set "BUILD_WINDOWS=false"
set "BUILD_LINUX=false"
set "BUILD_DARWIN=false"
set "OUTPUT_DIR=bin"

REM 解析命令行参数
:parseArgs
    if "%~1"=="" goto startBuild
    
    if /i "%~1"=="/h" (
        call :showHelp
        exit /b 0
    ) else if /i "%~1"=="/help" (
        call :showHelp
        exit /b 0
    ) else if /i "%~1"=="/o" (
        if /i "%~2"=="windows" (
            set "BUILD_WINDOWS=true"
        ) else if /i "%~2"=="linux" (
            set "BUILD_LINUX=true"
        ) else if /i "%~2"=="darwin" (
            set "BUILD_DARWIN=true"
        ) else (
            echo 错误: 不支持的操作系统: %~2
            exit /b 1
        )
        shift
    ) else if /i "%~1"=="/os" (
        if /i "%~2"=="windows" (
            set "BUILD_WINDOWS=true"
        ) else if /i "%~2"=="linux" (
            set "BUILD_LINUX=true"
        ) else if /i "%~2"=="darwin" (
            set "BUILD_DARWIN=true"
        ) else (
            echo 错误: 不支持的操作系统: %~2
            exit /b 1
        )
        shift
    ) else if /i "%~1"=="/a" (
        set "BUILD_ALL=true"
    ) else if /i "%~1"=="/all" (
        set "BUILD_ALL=true"
    ) else if /i "%~1"=="/v" (
        set "VERSION=%~2"
        shift
    ) else if /i "%~1"=="/version" (
        set "VERSION=%~2"
        shift
    ) else (
        echo 未知选项: %~1
        call :showHelp
        exit /b 1
    )
    
    shift
    goto parseArgs

:startBuild
    REM 如果指定了 /all，则编译所有支持的操作系统
    if "%BUILD_ALL%"=="true" (
        set "BUILD_WINDOWS=true"
        set "BUILD_LINUX=true"
        set "BUILD_DARWIN=true"
    )

    REM 如果没有指定任何操作系统，则仅编译Windows版
    if "%BUILD_WINDOWS%"=="false" if "%BUILD_LINUX%"=="false" if "%BUILD_DARWIN%"=="false" (
        set "BUILD_WINDOWS=true"
    )

    REM 创建输出目录
    if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

    REM 显示编译信息
    echo 开始编译运维工具箱 v%VERSION%
    echo 目标操作系统:
    if "%BUILD_WINDOWS%"=="true" echo - Windows
    if "%BUILD_LINUX%"=="true" echo - Linux
    if "%BUILD_DARWIN%"=="true" echo - macOS
    echo ----------------

    REM 编译Windows版本
    if "%BUILD_WINDOWS%"=="true" (
        call :build "windows" "amd64" "tuleaj-tools-windows-x64.exe" "%VERSION%"
        call :build "windows" "386" "tuleaj-tools-windows-x86.exe" "%VERSION%"
    )

    REM 编译Linux版本
    if "%BUILD_LINUX%"=="true" (
        call :build "linux" "amd64" "tuleaj-tools-linux-x64" "%VERSION%"
        call :build "linux" "386" "tuleaj-tools-linux-x86" "%VERSION%"
        call :build "linux" "arm64" "tuleaj-tools-linux-arm64" "%VERSION%"
    )

    REM 编译macOS版本
    if "%BUILD_DARWIN%"=="true" (
        call :build "darwin" "amd64" "tuleaj-tools-macos-x64" "%VERSION%"
        call :build "darwin" "arm64" "tuleaj-tools-macos-arm64" "%VERSION%"
    )

    echo ----------------
    echo 编译完成！
    exit /b 0

REM 编译函数
:build
    set "OS=%~1"
    set "ARCH=%~2"
    set "OUTPUT_NAME=%~3"
    set "VERSION=%~4"
    set "OUTPUT_PATH=%OUTPUT_DIR%\%OUTPUT_NAME%"
    
    echo 编译 %OS%/%ARCH% 版本...
    
    REM 设置环境变量
    set "CGO_ENABLED=0"
    set "GOOS=%OS%"
    set "GOARCH=%ARCH%"
    
    REM 编译
    go build -trimpath -ldflags "-s -w -X main.Version=%VERSION%" -o "%OUTPUT_PATH%" cmd/cli/main.go
    
    if %errorlevel% neq 0 (
        echo 编译失败: %OS%/%ARCH%
        exit /b 1
    ) else (
        echo 已成功创建: %OUTPUT_PATH%
    )
    
    exit /b 0 