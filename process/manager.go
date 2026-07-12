package process

import (
	"fmt"
	"sort"
	"strconv"

	ps "github.com/shirou/gopsutil/v3/process"
)

// ProcessInfo holds information about a running process.
type ProcessInfo struct {
	PID     int32
	Name    string
	CPU     float64
	Memory  float64
	Status  string
	CMDLine string
}

// ListProcesses returns all running processes sorted by CPU usage (descending).
func ListProcesses() ([]ProcessInfo, error) {
	processes, err := ps.Processes()
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %w", err)
	}

	var result []ProcessInfo
	for _, p := range processes {
		name, err := p.Name()
		if err != nil || name == "" {
			continue
		}

		cpu, _ := p.CPUPercent()
		mem, _ := p.MemoryPercent()
		status, _ := p.Status()
		cmdline, _ := p.Cmdline()

		statusStr := "running"
		if len(status) > 0 {
			statusStr = status[0]
		}

		result = append(result, ProcessInfo{
			PID:     p.Pid,
			Name:    name,
			CPU:     cpu,
			Memory:  float64(mem),
			Status:  statusStr,
			CMDLine: cmdline,
		})
	}

	// Sort by CPU usage descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].CPU > result[j].CPU
	})

	return result, nil
}

// GetChildProcesses returns all child PIDs of the given parent PID recursively.
func GetChildProcesses(pid int32) ([]int32, error) {
	processes, err := ps.Processes()
	if err != nil {
		return nil, err
	}

	var children []int32
	var findChildren func(parent int32)
	findChildren = func(parent int32) {
		for _, p := range processes {
			ppid, err := p.Ppid()
			if err != nil {
				continue
			}
			if ppid == parent {
				children = append(children, p.Pid)
				findChildren(p.Pid)
			}
		}
	}

	findChildren(pid)
	return children, nil
}

// KillProcess kills a process by PID. If killTree is true, it also kills all child processes.
func KillProcess(pid int32, killTree bool) error {
	if killTree {
		children, err := GetChildProcesses(pid)
		if err != nil {
			// Continue even if we can't get children
		} else {
			// Kill children first (deepest first)
			for i := len(children) - 1; i >= 0; i-- {
				killSingleProcess(children[i])
			}
		}
	}

	return killSingleProcess(pid)
}

func killSingleProcess(pid int32) error {
	proc, err := ps.NewProcess(pid)
	if err != nil {
		// Process not found = already dead, that's fine
		return nil
	}

	// Try SIGTERM first, then SIGKILL
	err = proc.Terminate()
	if err != nil {
		// If terminate fails, try kill
		err = proc.Kill()
		if err != nil {
			// Process might already be gone or permission denied
			// Check if process still exists - if not, it's already dead
			if exists, _ := ps.PidExists(pid); !exists {
				return nil
			}
			return fmt.Errorf("permission denied or unable to kill process %d", pid)
		}
	}

	return nil
}

// FormatBytes formats bytes to human-readable string.
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return strconv.FormatUint(bytes, 10) + " B"
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
