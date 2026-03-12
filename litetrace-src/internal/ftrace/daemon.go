package ftrace

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	PidFileDir  = "/var/run/litetrace"
	PidFileName = "litetrace.pid"
)

// GetPidFilePath 返回PID文件的完整路径
func GetPidFilePath() string {
	return filepath.Join(PidFileDir, PidFileName)
}

// WritePidFile 将指定PID写入文件
func WritePidFile() error {
	return WritePidFileWithPid(os.Getpid())
}

// WritePidFileWithPid 将指定PID写入文件
func WritePidFileWithPid(pid int) error {
	// 确保目录存在
	if err := os.MkdirAll(PidFileDir, 0755); err != nil {
		return fmt.Errorf("failed to create pid directory: %w", err)
	}

	pidFile := GetPidFilePath()

	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
		return fmt.Errorf("failed to write pid file: %w", err)
	}

	return nil
}

// ReadPidFile 读取PID文件中的进程ID
func ReadPidFile() (int, error) {
	pidFile := GetPidFilePath()
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("invalid pid file content: %w", err)
	}

	return pid, nil
}

// RemovePidFile 删除PID文件
func RemovePidFile() error {
	pidFile := GetPidFilePath()
	return os.Remove(pidFile)
}

// IsDaemonRunning 检查后台进程是否正在运行
func IsDaemonRunning() (bool, int, error) {
	pid, err := ReadPidFile()
	if err != nil {
		if os.IsNotExist(err) {
			return false, 0, nil
		}
		return false, 0, err
	}

	// 检查进程是否存在
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, 0, nil
	}

	// 发送信号0检查进程是否存在
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		// 进程不存在，清理PID文件
		RemovePidFile()
		return false, 0, nil
	}

	return true, pid, nil
}

// StopDaemon 停止后台进程
func StopDaemon() error {
	running, pid, err := IsDaemonRunning()
	if err != nil {
		return err
	}

	if !running {
		return fmt.Errorf("no daemon is running")
	}

	// 发送SIGTERM信号
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send signal: %w", err)
	}

	// 等待进程退出
	time.Sleep(500 * time.Millisecond)

	// 检查进程是否还在运行
	running, _, _ = IsDaemonRunning()
	if running {
		// 强制终止
		process.Signal(syscall.SIGKILL)
	}

	RemovePidFile()
	return nil
}

// StartBackgroundProcess 启动一个真正的后台进程
// 使用setsid创建新的会话，完全脱离终端
func StartBackgroundProcess(tracer, filter, duration string, outputPath string) error {
	// 检查是否已有后台进程在运行
	running, pid, _ := IsDaemonRunning()
	if running {
		return fmt.Errorf("daemon already running with PID %d", pid)
	}

	// 构建命令参数（使用background-daemon子命令）
	args := []string{"background-daemon", "--tracer", tracer}
	if filter != "" {
		args = append(args, "--filter", filter)
	}
	if duration != "" {
		args = append(args, "--duration", duration)
	}
	if outputPath != "" {
		args = append(args, "--output", outputPath)
	}

	// 获取当前可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// 创建命令
	cmd := exec.Command(execPath, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	// 设置进程组ID，使其脱离当前终端
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	// 启动进程
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start background process: %w", err)
	}

	// 等待一小段时间确保进程启动成功
	time.Sleep(100 * time.Millisecond)

	// 检查进程是否还在运行
	if cmd.Process != nil {
		// 发送信号0检查进程是否存在
		err := cmd.Process.Signal(syscall.Signal(0))
		if err != nil {
			return fmt.Errorf("background process failed to start")
		}

		// 写入PID文件（写入子进程的PID）
		if err := WritePidFileWithPid(cmd.Process.Pid); err != nil {
			// 如果写入PID文件失败，终止后台进程
			cmd.Process.Kill()
			return err
		}

		return nil
	}

	return fmt.Errorf("failed to start background process")
}
