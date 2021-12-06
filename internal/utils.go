package internal

import (
	"os"
	"os/exec"
)

func RunCommandAndLogToStderr(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
