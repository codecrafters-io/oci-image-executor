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
	// TODO: Find out which one of these is important!
	ioutil.WriteFile(filepath.Join(b.mountedRootFSPath, "sbin/init"), []byte(initScriptContents), 0777)

	return ioutil.WriteFile(filepath.Join(b.mountedRootFSPath, "init"), []byte(initScriptContents), 0777)
}

func (b *RootFSBuilder) createAndMountEmptyRootFS() error {
	b.rootFSPath = "/var/opt/oci-image-executor/root-fs"
	b.mountedRootFSPath = "/var/opt/oci-image-executor/root-fs-mount"

	os.RemoveAll(b.rootFSPath)
	os.RemoveAll(b.mountedRootFSPath)
	os.Mkdir(b.mountedRootFSPath, 0744)

	// fallocate is faster - see if this causes problems!

	var err error

	// if err = RunCommandAndLogToStderr("dd", "if=/dev/zero", fmt.Sprintf("of=%s", b.rootFSPath), "bs=1M", "count=1500"); err != nil {
	// 	return err
	// }

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
