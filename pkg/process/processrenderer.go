package process

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

// ProcessTreeRenderer 进程树渲染器接口
type ProcessTreeRenderer interface {
	Render(tree *ProcessTreeNode) error
}

// 颜色配置
var (
	rootColor    = color.New(color.FgGreen, color.Bold)
	pidColor     = color.New(color.FgCyan)
	memoryColor  = color.New(color.FgYellow)
	cpuColor     = color.New(color.FgMagenta)
	statusColor  = color.New(color.FgBlue)
	branchColor  = color.New(color.FgBlue)
	systemColor  = color.New(color.FgRed, color.Bold)
	specialColor = color.New(color.FgYellow, color.Bold)
	normalColor  = color.New(color.FgWhite)
	titleColor   = color.New(color.FgCyan, color.Bold).Add(color.Underline)
	errorColor   = color.New(color.FgRed, color.Bold)
)

// BasicProcessTreeRenderer 基本的进程树渲染器
type BasicProcessTreeRenderer struct {
	Writer     io.Writer // 输出目标
	ShowDetail bool      // 是否显示详细信息
	NoColor    bool      // 是否禁用颜色
	Title      string    // 标题
}

// NewBasicRenderer 创建基本渲染器
func NewBasicRenderer(showDetail bool, noColor bool) *BasicProcessTreeRenderer {
	return &BasicProcessTreeRenderer{
		Writer:     os.Stdout,
		ShowDetail: showDetail,
		NoColor:    noColor,
		Title:      "系统进程树:",
	}
}

// Render 渲染进程树
func (r *BasicProcessTreeRenderer) Render(tree *ProcessTreeNode) error {
	if tree == nil {
		return fmt.Errorf("进程树为空")
	}

	// 如果请求禁用颜色
	if r.NoColor {
		color.NoColor = true
	}

	// 打印标题
	if r.Title != "" {
		titleColor.Fprintln(r.Writer, r.Title)
		fmt.Fprintln(r.Writer) // 添加空行
	}

	// 创建表格
	table := tablewriter.NewWriter(r.Writer)
	if r.ShowDetail {
		table.SetHeader([]string{"进程", "PID", "内存 (MB)", "CPU (%)", "状态"})
	} else {
		table.SetHeader([]string{""})
	}

	// 设置表格样式
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	// 非详细模式下，设置空表头以隐藏表头
	if !r.ShowDetail {
		table.SetHeader([]string{""})
	}

	// 保存表格数据
	var rows [][]string

	// 使用遍历器处理树节点
	WalkProcessTree(tree, func(node *ProcessTreeNode, depth int, isLast bool, prefix string) {
		// 将节点格式化为行数据
		row := r.formatNodeAsRow(node, depth, isLast, prefix)
		if row != nil {
			rows = append(rows, row)
		}
	})

	// 添加行数据
	for _, row := range rows {
		table.Append(row)
	}

	// 渲染表格
	table.Render()

	return nil
}

// formatNodeAsRow 将节点格式化为表格行
func (r *BasicProcessTreeRenderer) formatNodeAsRow(node *ProcessTreeNode, depth int, isLast bool, prefix string) []string {
	if node == nil {
		return nil
	}

	proc := node.Process

	// 获取进程名称的格式化字符串
	var procName string
	if depth == 0 {
		// 根节点使用特殊符号
		procName = r.formatProcessName(proc, "●", "", isLast, node.IsSpecial)
	} else {
		procName = r.formatProcessName(proc, "", prefix, isLast, node.IsSpecial)
	}

	// 根据显示详情模式返回不同格式的行
	if r.ShowDetail {
		memUsage := fmt.Sprintf("%.2f", float64(proc.MemoryInfo.RSS)/1024/1024)
		cpuUsage := fmt.Sprintf("%.2f", proc.CPU)
		status := proc.Status

		// 为System Idle Process添加特殊状态
		if node.IsSpecial && proc.PID == 0 {
			status = "系统空闲"
		}

		return []string{
			procName,
			pidColor.Sprintf("%d", proc.PID),
			memoryColor.Sprintf("%s", memUsage),
			cpuColor.Sprintf("%s", cpuUsage),
			statusColor.Sprintf("%s", status),
		}
	} else {
		return []string{procName}
	}
}

// formatProcessName 格式化进程名称
func (r *BasicProcessTreeRenderer) formatProcessName(p ProcessInfo, symbol string, prefix string, isLast bool, isSpecial bool) string {
	var linePrefix string
	if symbol != "" {
		linePrefix = symbol
	} else if isLast {
		linePrefix = "└─"
	} else {
		linePrefix = "├─"
	}

	// 添加分支前缀
	branchStr := branchColor.Sprintf("%s%s", prefix, linePrefix)

	// 根据进程属性选择颜色和格式
	var nameWithPid string

	if isSpecial && p.PID == 0 {
		// System Idle Process 特殊标记
		nameWithPid = systemColor.Sprintf("%s (PID=%d) [特殊系统进程]", p.Name, p.PID)
	} else {
		// 其他进程根据类型和PID选择颜色
		var procNameColor *color.Color
		if p.PID <= 4 {
			procNameColor = systemColor
		} else if strings.Contains(strings.ToLower(p.Name), "svchost") {
			procNameColor = color.New(color.FgHiBlue)
		} else if strings.HasSuffix(strings.ToLower(p.Name), ".exe") {
			procNameColor = normalColor
		} else {
			procNameColor = color.New(color.FgHiWhite)
		}

		nameWithPid = procNameColor.Sprintf("%s (PID=%d)", p.Name, p.PID)
	}

	return fmt.Sprintf("%s %s", branchStr, nameWithPid)
}

// TableProcessTreeRenderer 表格式进程树渲染器
type TableProcessTreeRenderer struct {
	*BasicProcessTreeRenderer
}

// NewTableRenderer 创建表格渲染器
func NewTableRenderer(showDetail bool, noColor bool) *TableProcessTreeRenderer {
	return &TableProcessTreeRenderer{
		BasicProcessTreeRenderer: NewBasicRenderer(showDetail, noColor),
	}
}

// 其他自定义渲染器可以在此继续添加
