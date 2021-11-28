package main

import (
	"fmt"
	"os"
)

func main() {
	config := ParseConfig()
	config.ValidatePathsExist()

	rootFSBuilder := NewRootFSBuilder(config)
	rootFSPath, err := rootFSBuilder.Build()
	if err != nil {
		fmt.Println(err)
		os.Exit(11)
	}

	defer rootFSBuilder.Cleanup()

	fmt.Println("path")
	fmt.Println(rootFSBuilder.mountedRootFSPath)
	RunCommandAndLogToStderr("ls", rootFSBuilder.mountedRootFSPath)

	fmt.Printf("hello world, %s %s %s\n", config.imageConfigFilePath, config.imageTarFilePath, rootFSPath)
}
