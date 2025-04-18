# 运维工具箱

一个用Go语言编写的运维工具箱，提供多种网络诊断、进程管理、文件处理和文本工具，适用于Windows和Linux系统。

## 安装

### 从源代码编译

需要Go 1.23.4或更高版本。

```bash
# 克隆仓库
git clone https://github.com/1ajie1/toolbox.git
cd tool-box

# 获取依赖
go mod tidy

# 编译
go build -o tool cmd/cli/main.go
```

### 预编译二进制文件

可以从[发布页面](https://github.com/1ajie1/toolbox.git)下载预编译的二进制文件。

## 命令速查

```
toolbox
├── network     网络诊断工具集
│   ├── ping        执行Ping测试
│   ├── portscan    执行端口扫描
│   ├── dns         执行DNS查询
│   ├── ipinfo      获取IP地址信息
│   ├── speedtest   执行网络速度测试
│   ├── traceroute  执行路由跟踪
│   ├── cert        证书的检查与生成
│   └── sniff       执行网络抓包
│
├── process     进程管理工具
│   ├── list        列出系统进程
│   ├── tree        以树形结构显示进程
│   ├── info        显示进程详情
│   ├── kill        终止指定进程
│   └── children    列出指定进程的子进程
│
├── fs          文件系统工具集
│   ├── compress    压缩或解压缩文件
│   ├── find        搜索文件和目录
│   └── tree        显示目录结构
│
├── fmt         格式化数据文件或文本内容
│   └── fmt         格式化数据（JSON/XML/YAML）
│
├── text        文本处理工具
│   ├── grep        文本搜索
│   ├── replace     文本替换
│   └── filter      文本过滤
│
└── help        显示帮助信息
```