//go:build !windows

package daemon

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

func daemonSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}

func stopProcess(process *os.Process) error {
	return process.Signal(syscall.SIGTERM)
}

func platformProcessAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil || errors.Is(err, syscall.EPERM)
}

func processCommand(pid int) ([]byte, error) {
	return exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "command=").Output()
}
