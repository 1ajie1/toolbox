# 运维工具箱跨平台编译指南

本文档提供了如何在多个操作系统上编译和使用运维工具箱的说明。

## 前置条件

- Go 编程语言环境 (推荐 Go 1.18 或更高版本)
- Git (用于获取源代码)

## 获取源码

```bash
git clone <repository-url>
cd toolbox
```

## 跨平台编译

### 使用脚本自动编译

我们提供了两种脚本来简化编译过程：

#### Linux/macOS 上使用 `build.sh`

首先确保脚本有执行权限：

```bash
chmod +x build.sh
```

编译所有平台:

```bash
./build.sh --all
```

编译特定平台:

```bash
./build.sh --os windows  # 只编译Windows版本
./build.sh --os linux    # 只编译Linux版本
./build.sh --os darwin   # 只编译macOS版本
```

指定版本号:

```bash
./build.sh --all --version 1.2.3
```

#### Windows 上使用 `build.bat`

编译所有平台:

```cmd
build.bat /all
```

编译特定平台:

```cmd
build.bat /os windows  # 只编译Windows版本
build.bat /os linux    # 只编译Linux版本
build.bat /os darwin   # 只编译macOS版本
```

指定版本号:

```cmd
build.bat /all /version 1.2.3
```

### 手动编译

如果您需要手动编译，可以使用以下命令：

#### Windows 编译

```bash
# 在Linux/macOS上交叉编译Windows 64位可执行文件
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -o bin/tuleaj-tools-windows-x64.exe cmd/cli/main.go

# 在Linux/macOS上交叉编译Windows 32位可执行文件
CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -trimpath -o bin/tuleaj-tools-windows-x86.exe cmd/cli/main.go
```

在Windows上编译其他平台:

```cmd
rem 交叉编译Linux 64位可执行文件
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -trimpath -o bin/tuleaj-tools-linux-x64 cmd/cli/main.go
```

#### Linux 编译

```bash
# 编译Linux 64位可执行文件
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o bin/tuleaj-tools-linux-x64 cmd/cli/main.go

# 编译Linux ARM64可执行文件 (如用于树莓派等)
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -o bin/tuleaj-tools-linux-arm64 cmd/cli/main.go
```

#### macOS 编译

```bash
# 编译macOS Intel版
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -o bin/tuleaj-tools-macos-x64 cmd/cli/main.go

# 编译macOS Apple Silicon (M1/M2) 版
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -o bin/tuleaj-tools-macos-arm64 cmd/cli/main.go
```

## 平台特定注意事项

### Windows 注意事项

- 工具箱在Windows上依赖WMI接口获取一些系统信息。
- 某些网络工具需要管理员权限才能正常工作。
- 如果遇到"此应用无法在您的电脑上运行"的提示，可以尝试右键点击可执行文件，选择"属性"，然后在"安全"选项卡中选择"解除锁定"。

### Linux 注意事项

- 网络嗅探功能需要root权限才能正常工作。
- 为获得最佳体验，可以将可执行文件复制到`/usr/local/bin`目录以便全局访问:
  ```bash
  sudo cp bin/tuleaj-tools-linux-x64 /usr/local/bin/tuleaj-tools
  ```

### macOS 注意事项

- 首次运行时，macOS可能会显示"无法验证开发者"的警告。您可以在"系统偏好设置" > "安全性与隐私"中允许应用运行。
- 或者在终端使用以下命令解锁:
  ```bash
  xattr -d com.apple.quarantine bin/tuleaj-tools-macos-x64
  ```

## 故障排除

### Windows 编译问题

如果在Windows上编译Linux版本时出现WMI相关错误，请确保您的Go环境设置了`CGO_ENABLED=0`，这将禁用cgo，从而避免平台特定的依赖问题。

### 找不到包错误

如果编译时遇到"找不到包"的错误，请确保已安装所有依赖:

```bash
go mod tidy
```

## 支持

如有任何编译或使用问题，请提交Issue或联系我们的支持团队。 