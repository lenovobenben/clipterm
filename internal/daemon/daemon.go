package daemon

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const appDir = "clipterm"

type StartOptions struct {
	DebugHotkeys bool
	PathStyle    string
}

type Status struct {
	Running bool
	PID     int
}

func Start(ctx context.Context, options StartOptions) (Status, error) {
	status, err := CurrentStatus()
	if err != nil {
		return Status{}, err
	}
	if status.Running {
		return status, nil
	}

	executable, err := os.Executable()
	if err != nil {
		return Status{}, err
	}

	logDir, err := LogDir()
	if err != nil {
		return Status{}, err
	}
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return Status{}, err
	}

	stdout, err := os.OpenFile(filepath.Join(logDir, "daemon.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return Status{}, err
	}
	defer stdout.Close()

	stderr, err := os.OpenFile(filepath.Join(logDir, "daemon.err.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return Status{}, err
	}
	defer stderr.Close()

	args := []string{"daemon", "--foreground"}
	if options.DebugHotkeys {
		args = append(args, "--debug-hotkeys")
	}
	if options.PathStyle != "" {
		args = append(args, "--path-style", options.PathStyle)
	}

	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = nil
	cmd.SysProcAttr = daemonSysProcAttr()

	if err := cmd.Start(); err != nil {
		return Status{}, err
	}

	if err := writePID(cmd.Process.Pid); err != nil {
		_ = cmd.Process.Kill()
		return Status{}, err
	}

	return Status{Running: true, PID: cmd.Process.Pid}, nil
}

func Stop() error {
	status, err := CurrentStatus()
	if err != nil {
		return err
	}
	if !status.Running {
		return removePID()
	}

	process, err := os.FindProcess(status.PID)
	if err != nil {
		return err
	}
	if err := stopProcess(process); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}

	return removePID()
}

func CurrentStatus() (Status, error) {
	pid, err := readPID()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Status{}, nil
		}
		return Status{}, err
	}

	if pid <= 0 {
		_ = removePID()
		return Status{}, nil
	}

	if processAlive(pid) && processLooksLikeDaemon(pid) {
		return Status{Running: true, PID: pid}, nil
	}

	_ = removePID()
	return Status{}, nil
}

func LogDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Logs", appDir), nil
}

func PIDPath() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appDir, "daemon.pid"), nil
}

func readPID() (int, error) {
	path, err := PIDPath()
	if err != nil {
		return 0, err
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(body)))
	if err != nil {
		return 0, fmt.Errorf("invalid daemon pid file %s: %w", path, err)
	}
	return pid, nil
}

func writePID(pid int) error {
	path, err := PIDPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strconv.Itoa(pid)), 0o644)
}

func removePID() error {
	path, err := PIDPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func processAlive(pid int) bool {
	return platformProcessAlive(pid)
}

func processLooksLikeDaemon(pid int) bool {
	output, err := processCommand(pid)
	if err != nil {
		return false
	}
	return commandLooksLikeDaemon(string(output))
}

func commandLooksLikeDaemon(command string) bool {
	fields := strings.Fields(command)
	if len(fields) < 3 {
		return false
	}

	executable := filepath.Base(fields[0])
	if executable != "clipterm" && executable != "clipterm.exe" {
		return false
	}

	for i := 1; i < len(fields); i++ {
		if fields[i] == "daemon" && i+1 < len(fields) && fields[i+1] == "--foreground" {
			return true
		}
	}
	return false
}
