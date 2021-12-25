package internal

import (
	"errors"
	"fmt"
	"os"

	flags "github.com/jessevdk/go-flags"
)

type Config struct {
	Cmd                  []string
	ImageTarFilePath     string            `long:"image-tar" description:"Path to the image tar file" required:"true"`
	ImageConfigFilePath  string            `long:"image-config" description:"Path to the image config file" required:"true"`
	Volumes              map[string]string `long:"volume" description:"A file/directory to copy into the VM. Format: /path/to/host/file:/path/to/vm/file"`
	EnvironmentVariables []string          `long:"env" description:"Environment variables for the process. Format: VAR=value"`
	WorkingDirectory     string            `long:"working-dir" description:"Working directory for the process." default:"/"`
}

func ParseConfig() Config {
	var config Config

	remainingArgs, err := flags.Parse(&config)

	if err != nil {
		fmt.Println("Error parsing arguments.")
		os.Exit(11) // Helps differentiate between exit code from process and exit code from executor
	}

	if len(remainingArgs) == 0 || (remainingArgs[0] == "") {
		fmt.Println("Expected cmd arg to be provided")
	}

	config.Cmd = remainingArgs

	return config
}

func (c Config) ValidatePathsExist() {
	if _, err := os.Stat(c.ImageTarFilePath); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("image tar file '%s' does not exist\n", c.ImageTarFilePath)
		os.Exit(11) // Helps differentiate between exit code from process and exit code from executor
	}

	if _, err := os.Stat(c.ImageConfigFilePath); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("image config file '%s' does not exist\n", c.ImageTarFilePath)
		os.Exit(11) // Helps differentiate between exit code from process and exit code from executor
	}

	for hostPath, _ := range c.Volumes {
		if _, err := os.Stat(hostPath); errors.Is(err, os.ErrNotExist) {
			fmt.Println("volume doesn't exist", hostPath)
			os.Exit(11) // Helps differentiate between exit code from process and exit code from executor
		}
	}
}
