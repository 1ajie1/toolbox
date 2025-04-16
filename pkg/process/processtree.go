package process

import (
	"fmt"
	"sort"
)

// ProcessTreeNode 表示进程树节点
type ProcessTreeNode struct {
	Process   ProcessInfo        // 当前进程信息
	Children  []*ProcessTreeNode // 子进程列表
	IsSpecial bool               // 是否为特殊进程
}

// ProcessTreeOptions 表示构建进程树的选项
type ProcessTreeOptions struct {
	RootPID       int32  // 根进程ID，默认为0表示整个系统
	Filter        string // 进程名称过滤条件
	IncludeOrphan bool   // 是否包含孤立进程
}

// BuildProcessTree 构建进程树
func BuildProcessTree(processList []ProcessInfo, options ProcessTreeOptions) (*ProcessTreeNode, error) {
	if len(processList) == 0 {
		return nil, fmt.Errorf("进程列表为空")
	}

	// 构建进程ID到进程信息的映射
	pidMap := make(map[int32]ProcessInfo)
	for _, p := range processList {
		pidMap[p.PID] = p
	}

	// 构建父进程ID到子进程的映射
	childrenMap := make(map[int32][]ProcessInfo)
	for _, p := range processList {
		childrenMap[p.PPID] = append(childrenMap[p.PPID], p)
	}

	// 对每个父进程的子进程列表按PID排序
	for ppid, children := range childrenMap {
		sort.Slice(children, func(i, j int) bool {
			return children[i].PID < children[j].PID
		})
		childrenMap[ppid] = children
	}

	// 已处理进程的集合，避免循环引用
	visited := make(map[int32]bool)

	// 创建进程树
	var rootNode *ProcessTreeNode

	// 根据不同的根进程ID处理
	if options.RootPID == 0 {
		// 系统的整体进程树

		// 查找系统进程（PID=4）作为根节点
		systemProc, systemExists := pidMap[4]

		if !systemExists {
			// 如果找不到系统进程，使用第一个PID较小的进程作为根节点
			for _, p := range processList {
				if p.PID <= 4 && p.PID != 0 { // 排除System Idle Process
					systemProc = p
					systemExists = true
					break
				}
			}
		}

		if systemExists {
			// 使用实际的System进程作为根节点
			rootNode = &ProcessTreeNode{
				Process:  systemProc,
				Children: []*ProcessTreeNode{},
			}
			visited[systemProc.PID] = true
		} else {
			// 找不到合适的系统进程，创建一个虚拟根节点
			rootNode = &ProcessTreeNode{
				Process: ProcessInfo{
					PID:  0,
					Name: "System",
				},
				Children: []*ProcessTreeNode{},
			}
		}

		// 处理 System Idle Process (PID=0) 作为特殊进程
		if idleProc, exists := pidMap[0]; exists {
			idleNode := &ProcessTreeNode{
				Process:   idleProc,
				Children:  []*ProcessTreeNode{},
				IsSpecial: true,
			}
			visited[idleProc.PID] = true
			rootNode.Children = append(rootNode.Children, idleNode)
		}

		// 添加其他系统进程作为子节点
		buildChildNodes(rootNode, 4, childrenMap, pidMap, visited, options.Filter)

		// 处理属于PID=0的子进程
		if rootProcs, exists := childrenMap[0]; exists {
			for _, p := range rootProcs {
				if !visited[p.PID] && p.PID != 0 { // 排除System Idle Process自身
					childNode := createProcessNode(p, childrenMap, pidMap, visited, options.Filter)
					if childNode != nil {
						rootNode.Children = append(rootNode.Children, childNode)
					}
				}
			}
		}

		// 如果需要包含孤立进程
		if options.IncludeOrphan {
			orphanProcs := getOrphanProcs(processList, pidMap, visited)
			for _, proc := range orphanProcs {
				childNode := createProcessNode(proc, childrenMap, pidMap, visited, options.Filter)
				if childNode != nil {
					rootNode.Children = append(rootNode.Children, childNode)
				}
			}
		}
	} else {
		// 指定进程的子树
		proc, exists := pidMap[options.RootPID]
		if !exists {
			return nil, fmt.Errorf("未找到PID为 %d 的进程", options.RootPID)
		}

		rootNode = &ProcessTreeNode{
			Process:  proc,
			Children: []*ProcessTreeNode{},
		}
		visited[proc.PID] = true

		// 构建子节点
		buildChildNodes(rootNode, options.RootPID, childrenMap, pidMap, visited, options.Filter)
	}

	return rootNode, nil
}

// buildChildNodes 递归构建子节点
func buildChildNodes(node *ProcessTreeNode, pid int32, childrenMap map[int32][]ProcessInfo, pidMap map[int32]ProcessInfo, visited map[int32]bool, filter string) {
	children, exists := childrenMap[pid]
	if !exists || len(children) == 0 {
		return
	}

	// 过滤掉已访问的进程和特殊情况
	var filteredChildren []ProcessInfo
	for _, child := range children {
		// 排除PID=4的System进程作为System Idle Process的子进程（避免循环）
		if pid == 0 && child.PID == 4 {
			continue
		}
		// 排除已访问的进程
		if visited[child.PID] {
			continue
		}
		filteredChildren = append(filteredChildren, child)
	}

	// 处理过滤后的子进程
	for _, child := range filteredChildren {
		childNode := createProcessNode(child, childrenMap, pidMap, visited, filter)
		if childNode != nil {
			node.Children = append(node.Children, childNode)
		}
	}
}

// createProcessNode 创建单个进程节点
func createProcessNode(proc ProcessInfo, childrenMap map[int32][]ProcessInfo, pidMap map[int32]ProcessInfo, visited map[int32]bool, filter string) *ProcessTreeNode {
	// 检查是否已访问
	if visited[proc.PID] {
		return nil
	}

	// 标记为已访问
	visited[proc.PID] = true

	// 如果有过滤条件且当前进程不匹配
	if filter != "" && !containsIgnoreCase(proc.Name, filter) {
		// 看看子进程中是否有匹配的
		hasMatchingChild := hasMatchingChildren(proc.PID, childrenMap, filter, make(map[int32]bool))

		// 如果该进程及其所有子进程都不匹配过滤条件，则跳过
		if !hasMatchingChild {
			// 由于跳过，要移除已访问标记
			delete(visited, proc.PID)
			return nil
		}
	}

	// 创建新节点
	node := &ProcessTreeNode{
		Process:  proc,
		Children: []*ProcessTreeNode{},
	}

	// 递归处理子进程
	buildChildNodes(node, proc.PID, childrenMap, pidMap, visited, filter)

	return node
}

// hasMatchingChildren 检查是否有子进程匹配过滤条件
func hasMatchingChildren(pid int32, childrenMap map[int32][]ProcessInfo, filter string, visited map[int32]bool) bool {
	// 防止循环引用
	if visited[pid] {
		return false
	}
	visited[pid] = true

	children, exists := childrenMap[pid]
	if !exists || len(children) == 0 {
		return false
	}

	for _, child := range children {
		if containsIgnoreCase(child.Name, filter) {
			return true
		}

		// 递归检查子进程
		if hasMatchingChildren(child.PID, childrenMap, filter, visited) {
			return true
		}
	}

	return false
}

// TraverseProcessTree 遍历进程树节点，对每个节点执行回调函数
func TraverseProcessTree(node *ProcessTreeNode, depth int, isLast bool, prefix string, callback func(*ProcessTreeNode, int, bool, string)) {
	// 先处理当前节点
	callback(node, depth, isLast, prefix)

	// 处理子节点
	if len(node.Children) == 0 {
		return
	}

	// 确定子节点的前缀
	var childPrefix string
	if isLast {
		childPrefix = prefix + "    "
	} else {
		childPrefix = prefix + "│   "
	}

	// 递归处理所有子节点
	for i, child := range node.Children {
		isChildLast := i == len(node.Children)-1
		TraverseProcessTree(child, depth+1, isChildLast, childPrefix, callback)
	}
}

// WalkProcessTree 遍历进程树的便捷方法
func WalkProcessTree(root *ProcessTreeNode, callback func(*ProcessTreeNode, int, bool, string)) {
	TraverseProcessTree(root, 0, true, "", callback)
}

// FilterProcessTree 根据条件过滤进程树
func FilterProcessTree(tree *ProcessTreeNode, filterFunc func(*ProcessTreeNode) bool) *ProcessTreeNode {
	if tree == nil {
		return nil
	}

	// 如果当前节点不符合条件，返回nil
	if !filterFunc(tree) {
		return nil
	}

	// 创建新节点副本
	newNode := &ProcessTreeNode{
		Process:   tree.Process,
		IsSpecial: tree.IsSpecial,
		Children:  []*ProcessTreeNode{},
	}

	// 递归过滤子节点
	for _, child := range tree.Children {
		filteredChild := FilterProcessTree(child, filterFunc)
		if filteredChild != nil {
			newNode.Children = append(newNode.Children, filteredChild)
		}
	}

	return newNode
}

// getOrphanProcs 获取孤立进程（那些父进程不存在的进程）
func getOrphanProcs(processList []ProcessInfo, pidMap map[int32]ProcessInfo, visited map[int32]bool) []ProcessInfo {
	var orphanProcs []ProcessInfo
	for _, proc := range processList {
		if !visited[proc.PID] && proc.PID != 0 && proc.PID != 4 {
			// 检查父进程是否存在
			if _, exists := pidMap[proc.PPID]; !exists {
				// 这是一个孤立进程（父进程不存在）
				orphanProcs = append(orphanProcs, proc)
			}
		}
	}
	return orphanProcs
}
