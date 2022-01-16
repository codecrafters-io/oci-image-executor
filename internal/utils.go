package internal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func RunCommandAndLogToStderr(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	defer func() {
		fmt.Printf("Ran command %s %s\n", name, strings.Join(args, " "))
	}()

	return cmd.Run()
}
