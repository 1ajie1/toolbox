#!/bin/bash

# 错误处理
set -e

# 显示帮助信息
show_help() {
    echo "运维工具箱跨平台编译脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo "选项:"
    echo "  -h, --help            显示帮助信息"
    echo "  -o, --os OS_NAME      指定目标操作系统 (windows, linux, darwin)"
    echo "  -a, --all             编译所有支持的操作系统"
    echo "  -v, --version VERSION 指定版本号 (默认: 1.0.0)"
    echo ""
    echo "示例:"
    echo "  $0 --os windows       仅编译Windows版本"
    echo "  $0 --all              编译所有操作系统的版本"
    echo "  $0 --all --version 1.2.3   编译所有操作系统的1.2.3版本"
}

# 默认值
VERSION="1.0.0"
BUILD_ALL=false
BUILD_WINDOWS=false
BUILD_LINUX=false
BUILD_DARWIN=false
OUTPUT_DIR="bin"

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
        -h|--help)
            show_help
            exit 0
            ;;
        -o|--os)
            case "$2" in
                windows) BUILD_WINDOWS=true ;;
                linux) BUILD_LINUX=true ;;
                darwin) BUILD_DARWIN=true ;;
                *) echo "错误: 不支持的操作系统: $2"; exit 1 ;;
            esac
            shift 2
            ;;
        -a|--all)
            BUILD_ALL=true
            shift
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        *)
            echo "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
done

# 如果指定了 --all，则编译所有支持的操作系统
if [ "$BUILD_ALL" = true ]; then
    BUILD_WINDOWS=true
    BUILD_LINUX=true
    BUILD_DARWIN=true
fi

# 如果没有指定任何操作系统，则根据当前系统编译
if [ "$BUILD_WINDOWS" = false ] && [ "$BUILD_LINUX" = false ] && [ "$BUILD_DARWIN" = false ]; then
    case "$(uname -s)" in
        MINGW*|MSYS*|CYGWIN*) BUILD_WINDOWS=true ;;
        Linux*) BUILD_LINUX=true ;;
        Darwin*) BUILD_DARWIN=true ;;
        *) echo "无法检测当前操作系统，请指定目标操作系统"; exit 1 ;;
    esac
fi

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

# 编译函数
build() {
    local os=$1
    local arch=$2
    local output_name=$3
    local version=$4
    local output_path="${OUTPUT_DIR}/${output_name}"
    
    echo "编译 ${os}/${arch} 版本..."
    
    # 设置环境变量
    export CGO_ENABLED=0
    export GOOS=$os
    export GOARCH=$arch
    
    # 如果是Windows，添加.exe后缀
    if [ "$os" = "windows" ]; then
        output_path="${output_path}.exe"
    fi
    
    # 编译
    go build -trimpath -ldflags "-s -w -X main.Version=${version}" -o "$output_path" cmd/cli/main.go
    
    echo "已成功创建: $output_path"
}

# 显示编译信息
echo "开始编译运维工具箱 v${VERSION}"
echo "目标操作系统: $([ "$BUILD_WINDOWS" = true ] && echo -n "Windows "; [ "$BUILD_LINUX" = true ] && echo -n "Linux "; [ "$BUILD_DARWIN" = true ] && echo -n "macOS")"
echo "----------------"

# 执行编译
if [ "$BUILD_WINDOWS" = true ]; then
    build "windows" "amd64" "tuleaj-tools-windows-x64" "$VERSION"
    build "windows" "386" "tuleaj-tools-windows-x86" "$VERSION"
fi

if [ "$BUILD_LINUX" = true ]; then
    build "linux" "amd64" "tuleaj-tools-linux-x64" "$VERSION"
    build "linux" "386" "tuleaj-tools-linux-x86" "$VERSION"
    build "linux" "arm64" "tuleaj-tools-linux-arm64" "$VERSION"
fi

if [ "$BUILD_DARWIN" = true ]; then
    build "darwin" "amd64" "tuleaj-tools-macos-x64" "$VERSION"
    build "darwin" "arm64" "tuleaj-tools-macos-arm64" "$VERSION"
fi

echo "----------------"
echo "编译完成！" 