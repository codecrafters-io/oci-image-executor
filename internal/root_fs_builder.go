package internal

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

type RootFSBuilder struct {
	cmd                  []string
	environmentVariables []string
	imageTarFilePath     string
	imageConfig          ImageConfig
	rootFSPath           string
	mountedRootFSPath    string
	volumes              map[string]string
	workingDirectory     string
}

func NewRootFSBuilder(config Config) (*RootFSBuilder, error) {
	imageConfig, err := ImageConfigFromFile(config.ImageConfigFilePath)
	if err != nil {
		return &RootFSBuilder{}, err
	}

	return &RootFSBuilder{
		cmd:                  config.Cmd,
		environmentVariables: config.EnvironmentVariables,
		imageConfig:          imageConfig,
		imageTarFilePath:     config.ImageTarFilePath,
		volumes:              config.Volumes,
		workingDirectory:     config.WorkingDirectory,
	}, nil
}

func (b *RootFSBuilder) Build() (string, error) {
	if err := b.createAndMountEmptyRootFS(); err != nil {
		return "", err
	}

	fmt.Println("Created and mounted empty rootfs")

	if err := b.copyVolumesToMountedRootFS(); err != nil {
		return "", err
	}

	fmt.Println("Copied volumes to mounted rootfs")

	if err := b.addInitScriptToMountedRootFS(); err != nil {
		return "", err
	}

	fmt.Println("Added init script to empty rootfs")

	if err := b.unmountRootFS(); err != nil {
		return "", err
	}

	fmt.Println("Unmounted root fs")

	return b.rootFSPath, nil
}

func (b *RootFSBuilder) PrintDebugInfo() {
	fmt.Println("rootfsbuilder paths", b.mountedRootFSPath, b.rootFSPath)
}

func (b *RootFSBuilder) addInitScriptToMountedRootFS() error {
	initScriptTemplate := template.Must(
		template.New("init").Parse(`#!/bin/sh
set -e
mount proc /proc -t proc
mount sysfs /sys -t sysfs
haveged # generate entropy

echo "nameserver 8.8.8.8" > /etc/resolv.conf
echo "127.0.0.1       localhost" > /etc/hosts

{{range $key, $value := .ParsedImageEnv -}}
export {{$key}}="{{$value}}"
{{end}}

{{range $key, $value := .ParsedExecutorEnv -}}
export {{$key}}="{{$value}}"
{{end}}

cd {{.WorkingDirectory}}

exec {{range .Cmd}}"{{.}}" {{end}}`),
	)

	initScriptBuilder := strings.Builder{}
	if err := initScriptTemplate.Execute(&initScriptBuilder, map[string]interface{}{
		"ParsedExecutorEnv": parseEnv(b.environmentVariables),
		"ParsedImageEnv":    parseEnv(b.imageConfig.Env),
		"Cmd":               b.cmd,
		"WorkingDirectory":  b.workingDirectory,
	}); err != nil {
		panic(err)
	}

	initScriptContents := initScriptBuilder.String()
	return ioutil.WriteFile(filepath.Join(b.mountedRootFSPath, "init"), []byte(initScriptContents), 0777)
}

func parseEnv(env []string) map[string]string {
	result := make(map[string]string)

	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}

	return result
}

func (b *RootFSBuilder) copyVolumesToMountedRootFS() error {
	var wg sync.WaitGroup
	var globalErr error

	for hostPath, guestPath := range b.volumes {
		wg.Add(1)

		go func(hostPath string, guestPath string) {
			defer wg.Done()

			if err := RunCommandAndLogToStderr("rm", "-rf", path.Join(b.mountedRootFSPath, guestPath)); err != nil {
				globalErr = err
			}

			if err := RunCommandAndLogToStderr("cp", "-R", hostPath, path.Join(b.mountedRootFSPath, guestPath)); err != nil {
				globalErr = err
			}
		}(hostPath, guestPath)
	}

	wg.Wait()
	return globalErr
}

func (b *RootFSBuilder) createAndMountEmptyRootFS() error {
	rootFSFile, err := ioutil.TempFile("", "oci-image-executor-root-fs-")
	if err != nil {
		return err
	}

	mountedRootFSPath, err := ioutil.TempDir("", "oci-image-executor-root-fs-mnt-")
	if err != nil {
		return err
	}

	b.rootFSPath = rootFSFile.Name()
	b.mountedRootFSPath = mountedRootFSPath

	// TODO: Bring this down to 3G
	if err = RunCommandAndLogToStderr("fallocate", "-l", "10G", b.rootFSPath); err != nil {
		return err
	}

	if err = RunCommandAndLogToStderr("mkfs.ext4", b.rootFSPath); err != nil {
		return err
	}

	if err = RunCommandAndLogToStderr("mount", b.rootFSPath, b.mountedRootFSPath); err != nil {
		return err
	}

	if err = RunCommandAndLogToStderr("tar", "xf", b.imageTarFilePath, "-C", b.mountedRootFSPath); err != nil {
		return err
	}

	return nil
}

func (b *RootFSBuilder) unmountRootFS() error {
	if err := RunCommandAndLogToStderr("umount", b.mountedRootFSPath); err != nil {
		return err
	}

	return nil
}

func (b RootFSBuilder) Cleanup() {
	if b.rootFSPath != "" {
		fmt.Println("Removing rootfs", b.rootFSPath)
		err := os.Remove(b.rootFSPath)
		if err != nil {
			panic(err)
		}
	}

	if b.mountedRootFSPath != "" {
		fmt.Println("Removing mountedrootfs", b.mountedRootFSPath)
		err := os.RemoveAll(b.mountedRootFSPath)
		if err != nil {
			panic(err)
		}
	}
}
