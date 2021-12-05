package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type RootFSBuilder struct {
	imageTarFilePath  string
	rootFSPath        string
	mountedRootFSPath string
}

func NewRootFSBuilder(config Config) *RootFSBuilder {
	return &RootFSBuilder{
		imageTarFilePath: config.imageTarFilePath,
	}
}

func (b *RootFSBuilder) Build() (string, error) {
	if err := b.createAndMountEmptyRootFS(); err != nil {
		return "", err
	}

	if err := b.addInitScriptToRootFS(); err != nil {
		return "", err
	}

	if err := RunCommandAndLogToStderr("umount", b.mountedRootFSPath); err != nil {
		return "", err
	}

	return b.rootFSPath, nil
}

func (b *RootFSBuilder) addInitScriptToRootFS() error {
	initScriptContents := `#!/bin/sh
set -e
mount proc /proc -t proc
mount sysfs /sys -t sysfs
exec /bin/sh`

	return ioutil.WriteFile(filepath.Join(b.mountedRootFSPath, "init"), []byte(initScriptContents), 0777)
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

	os.RemoveAll(b.rootFSPath)
	os.RemoveAll(b.mountedRootFSPath)
	os.Mkdir(b.mountedRootFSPath, 0744)

	if err = RunCommandAndLogToStderr("fallocate", "-l", "1.5G", b.rootFSPath); err != nil {
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

func (b RootFSBuilder) Cleanup() {
	if b.rootFSPath != "" {
		os.Remove(b.rootFSPath)
	}

	if b.mountedRootFSPath != "" {
		os.RemoveAll(b.mountedRootFSPath)
	}
}
