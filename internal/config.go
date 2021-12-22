package internal

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	ImageTarFilePath    string
	ImageConfigFilePath string
	Volumes             map[string]string
}

func ParseConfig() Config {
	volumesStr := ""

	var config Config
	flag.StringVar(&config.ImageTarFilePath, "image-tar", "", "Path to the image tar file")
	flag.StringVar(&config.ImageConfigFilePath, "image-config", "", "Path to the image config file")
	flag.StringVar(&volumesStr, "Volumes", "", "Comma separated list of files/director to copy into VM")
	flag.Parse()

	if config.ImageTarFilePath == "" {
		fmt.Println("-image-tar not provided")
		os.Exit(11) // Helps differentiate between exit code from process and exit code from executor
	}

	if config.ImageConfigFilePath == "" {
		fmt.Println("image-config not provided")
		os.Exit(11) // Helps differentiate between exit code from process and exit code from executor
	}

	config.Volumes = map[string]string{}

	if volumesStr != "" {
		for _, volumeStr := range strings.Split(volumesStr, ",") {
			hostPath := strings.Split(volumeStr, ":")[0]
			guestPath := strings.Split(volumeStr, ":")[1]
			config.Volumes[hostPath] = guestPath
		}
	}

	return config
}

func (c Config) ValidatePathsExist() {
	if _, err := os.Stat(c.ImageTarFilePath); errors.Is(err, os.ErrNotExist) {
		fmt.Println("image tar file does not exist")
		os.Exit(11) // Helps differentiate between exit code from process and exit code from executor
	}

	if _, err := os.Stat(c.ImageConfigFilePath); errors.Is(err, os.ErrNotExist) {
		fmt.Println("image config file does not exist")
		os.Exit(11) // Helps differentiate between exit code from process and exit code from executor
	}

	for hostPath, _ := range c.Volumes {
		if _, err := os.Stat(hostPath); errors.Is(err, os.ErrNotExist) {
			fmt.Println("volume doesn't exist", hostPath)
			os.Exit(11) // Helps differentiate between exit code from process and exit code from executor
		}
	}
}
