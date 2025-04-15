# 运维工具箱

一个用Go语言编写的运维工具箱，提供多种网络诊断、数据处理和文本处理功能。

## 功能特性

### 网络诊断工具

- **Ping测试** - 测试网络连通性和延迟
  - 支持自定义ping次数和间隔
  - 显示往返时间统计

- **端口扫描** - 检测主机开放的端口
  - 支持TCP和UDP扫描
  - 可自定义端口范围
  - 显示端口状态和服务

- **DNS查询** - 查询域名的DNS记录
  - 支持A, AAAA, MX, NS, TXT等多种记录类型
  - 可指定使用的DNS服务器

- **路由跟踪** - 追踪网络数据包的路由路径
  - 显示每一跳的IP地址和延迟
  - 支持主机名解析

- **网络速度测试** - 测试网络下载和上传速度
  - 支持选择测试服务器
  - 显示详细的带宽和延迟信息

- **IP信息查询** - 查询IP地址的地理位置和相关信息
  - 显示国家、城市、ISP等信息
  - 显示地理坐标

- **网络抓包** - 捕获和分析网络数据包
  - 支持过滤器语法
  - 可查看数据包详情
  - 可保存捕获结果为pcap格式

### 数据处理工具

- **格式化工具** - 处理结构化数据
  - 支持JSON、XML、YAML格式
  - 支持美化显示和压缩
  - 彩色输出代码
  - 支持文件输入输出和字符串处理

### 文本处理工具

- **文本搜索 (grep)** - 在文件中搜索文本
  - 支持正则表达式
  - 显示行号
  - 高亮匹配内容
  - 递归搜索目录
  - 支持文件名过滤
  - 支持上下文显示

- **文本替换 (replace)** - 替换文件中的文本
  - 支持正则表达式
  - 全局替换或单行替换
  - 可原地修改文件或输出到标准输出
  - 支持创建备份

- **文本过滤 (filter)** - 根据条件筛选文本行
  - 支持类似awk的过滤表达式
  - 自定义输出格式
  - 自定义字段分隔符

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

## 使用示例

### 网络诊断

```bash
# Ping测试
tool ping example.com

# 端口扫描
tool portscan example.com --start-port 80 --end-port 100

# DNS查询
tool dns example.com --type mx

# 路由跟踪
tool traceroute example.com

# 速度测试
tool speedtest

# IP信息查询
tool ipinfo 8.8.8.8

# 网络抓包
tool sniff eth0 --filter "tcp and port 80"
tool sniff --list-interfaces
```

### 数据处理

```bash
# 格式化JSON
tool fmt data.json --pretty

# 压缩XML
tool fmt config.xml --compact

# 格式化YAML
tool fmt data.yaml --pretty

# 处理字符串
tool fmt -s '{"name":"John"}' --format json --pretty
```

### 文本处理

```bash
# 文本搜索
tool text grep "error" log.txt
tool text grep -r "func" ./src      # 递归搜索
tool text grep -i "pattern" file.txt # 忽略大小写

# 文本替换
tool text replace "old" "new" file.txt
tool text replace -I "error" "warning" log.txt  # 原地修改文件

# 文本过滤
tool text filter '$1 > 100' data.txt
tool text filter -F, '$2 == "ERROR"' log.csv
```

## 许可证

MIT

## 贡献

欢迎提交问题报告和功能建议。如果您想贡献代码，请先开issue讨论您想要更改的内容。

## 联系方式

如有问题，请联系：your-email@example.com 