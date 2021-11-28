package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

type Config struct {
	imageTarFilePath    string
	imageConfigFilePath string
}

func ParseConfig() Config {
	var config Config
	flag.StringVar(&config.imageTarFilePath, "image-tar", "", "Path to the image tar file")
	flag.StringVar(&config.imageConfigFilePath, "image-config", "", "Path to the image config file")
	flag.Parse()

	if config.imageTarFilePath == "" {
		fmt.Println("-image-tar not provided")
		os.Exit(11) // Helps differentiate between exit code from process and exit code from executor
	}

	if config.imageConfigFilePath == "" {
		fmt.Println("image-config not provided")
		os.Exit(11) // Helps differentiate between exit code from process and exit code from executor
	}

	return config
}

func (c Config) ValidatePathsExist() {
	if _, err := os.Stat(c.imageTarFilePath); errors.Is(err, os.ErrNotExist) {
		fmt.Println("image tar file does not exist")
		os.Exit(11) // Helps differentiate between exit code from process and exit code from executor
	}

	if _, err := os.Stat(c.imageConfigFilePath); errors.Is(err, os.ErrNotExist) {
		fmt.Println("image config file does not exist")
		os.Exit(11) // Helps differentiate between exit code from process and exit code from executor
	}
}
