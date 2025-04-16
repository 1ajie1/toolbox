package process

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// ProcessInfo 表示进程信息
type ProcessInfo struct {
	PID        int32     // 进程ID
	PPID       int32     // 父进程ID
	Name       string    // 进程名称
	Executable string    // 可执行文件路径
	Username   string    // 用户名
	Status     string    // 状态
	CreateTime time.Time // 创建时间
	CPU        float64   // CPU使用率
	Memory     float32   // 内存使用率(百分比)
	MemoryInfo struct {
		RSS  uint64 // 常驻集大小(RSS)，单位字节
		VMS  uint64 // 虚拟内存大小，单位字节
		Swap uint64 // 交换空间大小，单位字节
	} // 内存使用详情
	CmdLine   []string // 命令行
	Threads   int32    // 线程数
	OpenFiles []string // 打开的文件
}

// getNumWorkers 根据系统CPU核心数和进程数量计算最优的工作线程数
// 注意：尽管此函数会根据CPU核心数计算建议的工作线程数，
// 但在进程处理相关函数中，我们实际上固定使用2个工作线程，
// 因为经过测试，使用太多线程会导致API调用冲突和数据丢失。
// 这个函数保留作为参考，未来可能会根据需要调整并发策略。
func getNumWorkers(processCount int) int {
	// 获取系统CPU核心数
	numCPU := runtime.NumCPU()

	// 默认使用CPU核心数作为工作线程数
	// 但设置上限为4个线程，避免过多的线程导致数据争用和一致性问题
	// 注意：处理进程信息时，过高的并发度会导致某些系统API调用失败，从而丢失进程信息
	workers := numCPU
	if workers > 4 {
		workers = 4
	}

	// 确保工作线程数不超过进程数
	if workers > processCount {
		workers = processCount
	}

	// 至少使用1个工作线程
	if workers < 1 {
		workers = 1
	}

	return workers
}

// GetProcessList 获取系统中的进程列表
func GetProcessList() ([]ProcessInfo, error) {
	// 获取所有进程
	processes, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("获取进程列表失败: %v", err)
	}

	// 设置较低的并发数，确保稳定性
	numWorkers := getNumWorkers(len(processes))

	// 创建结果切片并预分配空间
	result := make([]ProcessInfo, 0, len(processes))

	// 使用互斥锁保护结果切片
	var mu sync.Mutex

	// 使用WaitGroup等待所有工作线程完成
	var wg sync.WaitGroup

	// 平均分配任务
	chunkSize := (len(processes) + numWorkers - 1) / numWorkers

	// 启动工作线程
	for i := 0; i < numWorkers; i++ {
		// 计算每个工作线程的任务范围
		start := i * chunkSize
		end := start + chunkSize
		if end > len(processes) {
			end = len(processes)
		}

		// 确保不处理空切片
		if start >= len(processes) {
			break
		}

		// 增加等待计数
		wg.Add(1)

		// 启动工作线程
		go func(procs []*process.Process) {
			// 确保完成时减少等待计数
			defer wg.Done()

			// 本地处理结果
			localResults := make([]ProcessInfo, 0, len(procs))

			// 处理分配的进程
			for _, p := range procs {
				// 创建进程信息对象，先设置PID
				info := ProcessInfo{
					PID: p.Pid,
				}

				// 特殊处理系统进程
				if info.PID == 0 {
					// System Idle Process (PID 0)
					info.Name = "System Idle Process"
					info.PPID = 0
					localResults = append(localResults, info)
					continue
				} else if info.PID == 4 {
					// System (PID 4)
					info.Name = "System"
					info.PPID = 0
					localResults = append(localResults, info)
					continue
				}

				// 获取进程名称
				name, err := p.Name()
				if err == nil && name != "" {
					info.Name = name
				} else {
					// 尝试使用备用方法获取名称
					if exe, err := p.Exe(); err == nil && exe != "" {
						info.Name = filepath.Base(exe)
					} else if cmdLine, err := p.Cmdline(); err == nil && cmdLine != "" {
						cmdParts := strings.Fields(cmdLine)
						if len(cmdParts) > 0 {
							info.Name = filepath.Base(cmdParts[0])
						}
					} else {
						// 如果名称获取失败，而且不是已知的系统进程，尝试确定是否为系统进程
						// 如果是普通用户进程但无法获取名称，则可能跳过
						if info.PID < 100 {
							// 可能是系统进程，使用一个特殊标识
							info.Name = fmt.Sprintf("System Process (%d)", info.PID)
						} else {
							// 跳过无法识别的普通进程
							continue
						}
					}
				}

				// 获取父进程ID
				if ppid, err := p.Ppid(); err == nil {
					info.PPID = ppid
				}

				// 获取用户名
				if username, err := p.Username(); err == nil {
					info.Username = username
				}

				// 获取CPU使用率
				if cpu, err := p.CPUPercent(); err == nil {
					info.CPU = cpu
				}

				// 获取内存使用率
				if memPercent, err := p.MemoryPercent(); err == nil {
					info.Memory = memPercent
				}

				// 获取命令行
				if cmdline, err := p.CmdlineSlice(); err == nil && len(cmdline) > 0 {
					info.CmdLine = cmdline
				} else if fullCmd, err := p.Cmdline(); err == nil && fullCmd != "" {
					info.CmdLine = strings.Fields(fullCmd)
				}

				// 添加到本地结果列表
				localResults = append(localResults, info)
			}

			// 合并到全局结果
			mu.Lock()
			result = append(result, localResults...)
			mu.Unlock()

		}(processes[start:end])
	}

	// 等待所有工作线程完成
	wg.Wait()

	// 按PID排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].PID < result[j].PID
	})

	return result, nil
}

// GetProcessByPID 通过PID获取特定进程信息
func GetProcessByPID(pid int32) (ProcessInfo, error) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return ProcessInfo{}, fmt.Errorf("进程不存在或无法访问: %v", err)
	}

	// 创建进程信息
	info := ProcessInfo{
		PID: p.Pid,
	}

	// 获取进程详细信息
	// 获取名称
	if name, err := p.Name(); err == nil {
		info.Name = name
	} else {
		// 尝试使用备用方法获取名称
		if exe, err := p.Exe(); err == nil && exe != "" {
			info.Name = filepath.Base(exe)
		} else if cmdLine, err := p.Cmdline(); err == nil && cmdLine != "" {
			cmdParts := strings.Fields(cmdLine)
			if len(cmdParts) > 0 {
				info.Name = filepath.Base(cmdParts[0])
			}
		}
	}

	// 获取父进程ID
	if ppid, err := p.Ppid(); err == nil {
		info.PPID = ppid
	}

	// 获取可执行文件路径
	if exe, err := p.Exe(); err == nil {
		info.Executable = exe
	}

	// 获取用户名
	if username, err := p.Username(); err == nil {
		info.Username = username
	}

	// 获取状态
	if status, err := p.Status(); err == nil && len(status) > 0 {
		info.Status = status[0]
	}

	// 获取创建时间
	if createTime, err := p.CreateTime(); err == nil {
		info.CreateTime = time.Unix(createTime/1000, 0)
	}

	// 获取CPU使用率
	if cpu, err := p.CPUPercent(); err == nil {
		info.CPU = cpu
	}

	// 获取内存使用率
	if memPercent, err := p.MemoryPercent(); err == nil {
		info.Memory = memPercent
	}

	// 获取内存使用详情
	if memInfo, err := p.MemoryInfo(); err == nil && memInfo != nil {
		info.MemoryInfo.RSS = memInfo.RSS
		info.MemoryInfo.VMS = memInfo.VMS
		info.MemoryInfo.Swap = memInfo.Swap
	}

	// 获取命令行
	if cmdline, err := p.CmdlineSlice(); err == nil {
		info.CmdLine = cmdline
	} else if fullCmd, err := p.Cmdline(); err == nil && fullCmd != "" {
		info.CmdLine = strings.Fields(fullCmd)
	}

	// 获取线程数
	if threadCount, err := p.NumThreads(); err == nil {
		info.Threads = threadCount
	}

	// 获取打开的文件
	if openFiles, err := p.OpenFiles(); err == nil {
		for _, f := range openFiles {
			info.OpenFiles = append(info.OpenFiles, f.Path)
		}
	}

	return info, nil
}

// KillProcess 结束指定PID的进程
func KillProcess(pid int32) error {
	p, err := process.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("找不到进程 PID=%d: %v", pid, err)
	}

	// 尝试正常终止进程
	if err := p.Terminate(); err != nil {
		// 如果正常终止失败，尝试强制结束
		if killErr := p.Kill(); killErr != nil {
			return fmt.Errorf("无法终止进程 PID=%d: %v", pid, killErr)
		}
	}

	return nil
}

// GetChildProcesses 获取指定PID的所有子进程
func GetChildProcesses(pid int32) ([]ProcessInfo, error) {
	allProcesses, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("获取进程列表失败: %v", err)
	}

	// 先过滤出所有可能的子进程，减少需要并发处理的数量
	var possibleChildren []*process.Process
	for _, p := range allProcesses {
		// 获取父进程ID
		ppid, err := p.Ppid()
		if err != nil {
			continue
		}

		// 如果父进程ID匹配，则为子进程
		if ppid == pid {
			possibleChildren = append(possibleChildren, p)
		}
	}

	// 如果没有找到子进程，直接返回
	if len(possibleChildren) == 0 {
		return []ProcessInfo{}, nil
	}

	// 设置较低的并发数，确保稳定性
	numWorkers := 2 // 固定使用2个工作线程，降低并发导致的问题
	if len(possibleChildren) < numWorkers {
		numWorkers = len(possibleChildren)
	}

	// 创建结果切片
	result := make([]ProcessInfo, 0, len(possibleChildren))

	// 使用互斥锁保护结果切片
	var mu sync.Mutex

	// 使用WaitGroup等待所有工作线程完成
	var wg sync.WaitGroup

	// 平均分配任务
	chunkSize := (len(possibleChildren) + numWorkers - 1) / numWorkers

	// 启动工作线程
	for i := 0; i < numWorkers; i++ {
		// 计算每个工作线程的任务范围
		start := i * chunkSize
		end := start + chunkSize
		if end > len(possibleChildren) {
			end = len(possibleChildren)
		}

		// 确保不处理空切片
		if start >= len(possibleChildren) {
			break
		}

		// 增加等待计数
		wg.Add(1)

		// 启动工作线程
		go func(procs []*process.Process) {
			// 确保完成时减少等待计数
			defer wg.Done()

			// 本地处理结果
			localResults := make([]ProcessInfo, 0, len(procs))

			// 处理分配的进程
			for _, p := range procs {
				// 创建进程信息
				childPid := p.Pid

				// 特殊处理系统进程（很少有系统进程是子进程，但以防万一）
				var procName string
				var isSystemProcess bool = false

				if childPid == 0 {
					procName = "System Idle Process"
					isSystemProcess = true
				} else if childPid == 4 {
					procName = "System"
					isSystemProcess = true
				} else {
					// 获取进程名称
					var err error
					procName, err = p.Name()
					if err != nil || procName == "" {
						// 尝试使用备用方法获取名称
						if exe, err := p.Exe(); err == nil && exe != "" {
							procName = filepath.Base(exe)
						} else if cmdLine, err := p.Cmdline(); err == nil && cmdLine != "" {
							cmdParts := strings.Fields(cmdLine)
							if len(cmdParts) > 0 {
								procName = filepath.Base(cmdParts[0])
							}
						} else {
							// 如果是可能的系统进程但无法获取名称
							if childPid < 100 {
								procName = fmt.Sprintf("System Process (%d)", childPid)
								isSystemProcess = true
							} else {
								// 如果无法获取名称，使用一个默认名称
								procName = fmt.Sprintf("Process (%d)", childPid)
							}
						}
					}
				}

				info := ProcessInfo{
					PID:  childPid,
					PPID: pid, // 已知父进程ID
					Name: procName,
				}

				// 如果是已知的系统进程，可能不需要获取更多信息
				if isSystemProcess {
					localResults = append(localResults, info)
					continue
				}

				// 获取可执行文件路径
				if exe, err := p.Exe(); err == nil {
					info.Executable = exe
				}

				// 获取用户名
				if username, err := p.Username(); err == nil {
					info.Username = username
				}

				// 获取状态
				if status, err := p.Status(); err == nil && len(status) > 0 {
					info.Status = status[0]
				}

				// 获取创建时间
				if createTime, err := p.CreateTime(); err == nil {
					info.CreateTime = time.Unix(createTime/1000, 0)
				}

				// 获取CPU使用率
				if cpu, err := p.CPUPercent(); err == nil {
					info.CPU = cpu
				}

				// 获取内存使用率
				if memPercent, err := p.MemoryPercent(); err == nil {
					info.Memory = memPercent
				}

				// 获取内存使用详情
				if memInfo, err := p.MemoryInfo(); err == nil && memInfo != nil {
					info.MemoryInfo.RSS = memInfo.RSS
					info.MemoryInfo.VMS = memInfo.VMS
					info.MemoryInfo.Swap = memInfo.Swap
				}

				// 获取命令行
				if cmdline, err := p.CmdlineSlice(); err == nil && len(cmdline) > 0 {
					info.CmdLine = cmdline
				} else if fullCmd, err := p.Cmdline(); err == nil && fullCmd != "" {
					info.CmdLine = strings.Fields(fullCmd)
				}

				// 获取线程数
				if threadCount, err := p.NumThreads(); err == nil {
					info.Threads = threadCount
				}

				// 获取打开的文件
				if openFiles, err := p.OpenFiles(); err == nil {
					for _, f := range openFiles {
						info.OpenFiles = append(info.OpenFiles, f.Path)
					}
				}

				localResults = append(localResults, info)
			}

			// 合并到全局结果
			mu.Lock()
			result = append(result, localResults...)
			mu.Unlock()

		}(possibleChildren[start:end])
	}

	// 等待所有工作线程完成
	wg.Wait()

	// 按PID排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].PID < result[j].PID
	})

	return result, nil
}

// StartProcess 启动新进程
func StartProcess(executable string, args []string) (int32, error) {
	// 使用os/exec包启动进程
	proc, err := os.StartProcess(
		executable,
		append([]string{executable}, args...),
		&os.ProcAttr{
			Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		},
	)
	if err != nil {
		return 0, fmt.Errorf("启动进程失败: %v", err)
	}

	// 返回新进程的PID
	return int32(proc.Pid), nil
}

// FilterProcessesByName 根据进程名称筛选进程
func FilterProcessesByName(name string) ([]ProcessInfo, error) {
	// 获取所有进程
	processes, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("获取进程列表失败: %v", err)
	}

	// 设置较低的并发数，确保稳定性
	numWorkers := 2 // 固定使用2个工作线程，降低并发导致的问题

	// 创建结果切片
	result := make([]ProcessInfo, 0)

	// 使用互斥锁保护结果切片
	var mu sync.Mutex

	// 使用WaitGroup等待所有工作线程完成
	var wg sync.WaitGroup

	// 平均分配任务
	chunkSize := (len(processes) + numWorkers - 1) / numWorkers

	// 启动工作线程
	for i := 0; i < numWorkers; i++ {
		// 计算每个工作线程的任务范围
		start := i * chunkSize
		end := start + chunkSize
		if end > len(processes) {
			end = len(processes)
		}

		// 确保不处理空切片
		if start >= len(processes) {
			break
		}

		// 增加等待计数
		wg.Add(1)

		// 启动工作线程
		go func(procs []*process.Process) {
			// 确保完成时减少等待计数
			defer wg.Done()

			// 本地处理结果
			localResults := make([]ProcessInfo, 0)

			// 处理分配的进程
			for _, p := range procs {
				// 保存PID
				pid := p.Pid

				// 特殊处理系统进程
				var procName string
				var isSystemProcess bool = false

				if pid == 0 {
					procName = "System Idle Process"
					isSystemProcess = true
				} else if pid == 4 {
					procName = "System"
					isSystemProcess = true
				} else {
					// 获取进程名称
					var err error
					procName, err = p.Name()
					if err != nil || procName == "" {
						// 尝试使用备用方法获取名称
						if exe, err := p.Exe(); err == nil && exe != "" {
							procName = filepath.Base(exe)
						} else if cmdLine, err := p.Cmdline(); err == nil && cmdLine != "" {
							cmdParts := strings.Fields(cmdLine)
							if len(cmdParts) > 0 {
								procName = filepath.Base(cmdParts[0])
							}
						} else {
							// 如果是可能的系统进程但无法获取名称
							if pid < 100 {
								procName = fmt.Sprintf("System Process (%d)", pid)
								isSystemProcess = true
							} else {
								// 跳过无法识别的普通进程
								continue
							}
						}
					}
				}

				// 检查名称是否匹配
				if containsIgnoreCase(procName, name) || (name == "system" && isSystemProcess) {
					// 创建进程信息
					info := ProcessInfo{
						PID:  pid,
						Name: procName,
					}

					// 如果是已知的系统进程，设置特定值
					if isSystemProcess {
						info.PPID = 0
						localResults = append(localResults, info)
						continue
					}

					// 获取父进程ID
					if ppid, err := p.Ppid(); err == nil {
						info.PPID = ppid
					}

					// 获取可执行文件路径
					if exe, err := p.Exe(); err == nil {
						info.Executable = exe
					}

					// 获取用户名
					if username, err := p.Username(); err == nil {
						info.Username = username
					}

					// 获取CPU使用率
					if cpu, err := p.CPUPercent(); err == nil {
						info.CPU = cpu
					}

					// 获取内存使用率
					if memPercent, err := p.MemoryPercent(); err == nil {
						info.Memory = memPercent
					}

					// 获取内存使用详情
					if memInfo, err := p.MemoryInfo(); err == nil && memInfo != nil {
						info.MemoryInfo.RSS = memInfo.RSS
						info.MemoryInfo.VMS = memInfo.VMS
						info.MemoryInfo.Swap = memInfo.Swap
					}

					// 获取命令行
					if cmdline, err := p.CmdlineSlice(); err == nil && len(cmdline) > 0 {
						info.CmdLine = cmdline
					} else if fullCmd, err := p.Cmdline(); err == nil && fullCmd != "" {
						info.CmdLine = strings.Fields(fullCmd)
					}

					// 获取线程数
					if threadCount, err := p.NumThreads(); err == nil {
						info.Threads = threadCount
					}

					localResults = append(localResults, info)
				}
			}

			// 合并到全局结果
			mu.Lock()
			result = append(result, localResults...)
			mu.Unlock()

		}(processes[start:end])
	}

	// 等待所有工作线程完成
	wg.Wait()

	// 按PID排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].PID < result[j].PID
	})

	return result, nil
}

// 不区分大小写的子字符串检查
func containsIgnoreCase(s, substr string) bool {
	s, substr = strings.ToLower(s), strings.ToLower(substr)
	return strings.Contains(s, substr)
}
