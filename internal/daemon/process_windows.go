//go:build windows

package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const createNewProcessGroup = 0x00000200

func daemonSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: createNewProcessGroup}
}

func stopProcess(process *os.Process) error {
	return process.Kill()
}

func platformProcessAlive(pid int) bool {
	output, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), fmt.Sprintf("%d", pid))
}

func processCommand(pid int) ([]byte, error) {
	return exec.Command(
		"powershell.exe",
		"-NoProfile",
		"-Command",
		fmt.Sprintf("(Get-CimInstance Win32_Process -Filter 'ProcessId=%d').CommandLine", pid),
	).Output()
}
