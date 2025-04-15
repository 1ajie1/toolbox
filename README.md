# 运维工具箱

一个用Go语言编写的运维工具箱，提供命令行和图形界面两种使用方式。

## 功能特性

当前版本包含以下网络诊断工具：

- **Ping测试**：测试网络连通性和延迟
- **端口扫描**：检测主机开放的端口
- **DNS查询**：查询域名的DNS记录
- **路由跟踪**：追踪网络数据包的路由路径
- **网络速度测试**：测试网络下载和上传速度
- **IP信息查询**：查询IP地址的地理位置和相关信息

## 安装

### 从源代码编译

需要Go 1.19或更高版本。

```bash
# 克隆仓库
git clone https://github.com/yourusername/tool-box.git
cd tool-box

# 获取依赖
go mod tidy

# 编译命令行版本
go build -o netdiag-cli cmd/cli/main.go

# 编译GUI版本
go build -o netdiag-gui cmd/gui/main.go
```

### 预编译二进制文件

可以从[发布页面](https://github.com/yourusername/tool-box/releases)下载预编译的二进制文件。


### 图形界面版本

直接运行可执行文件 `netdiag-gui`，然后通过界面操作。

## 构建依赖

- Go 1.19+
- [Fyne](https://fyne.io/) (GUI版本)
- [Cobra](https://github.com/spf13/cobra) (命令行版本)
- [fatih/color](https://github.com/fatih/color) (命令行彩色输出)

## 许可证

MIT

## 贡献

欢迎提交问题报告和功能建议。如果您想贡献代码，请先开issue讨论您想要更改的内容。

## 联系方式

如有问题，请联系：your-email@example.com 