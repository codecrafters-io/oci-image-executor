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

	if err := b.copyImageContentsIntoRootFS(); err != nil {
		return "", err
	}

	if err := b.addInitScriptToRootFS(); err != nil {
		return "", err
	}

	return b.rootFSPath, nil
}

func (b *RootFSBuilder) addInitScriptToRootFS() error {
	initScriptContents := `
#!/bin/sh
echo "hey"
sleep 1000
`
	return ioutil.WriteFile(filepath.Join(b.mountedRootFSPath, "/sbin/init"), []byte(initScriptContents), 0777)
}

func (b *RootFSBuilder) copyImageContentsIntoRootFS() error {
	return RunCommandAndLogToStderr("tar", "xf", b.imageTarFilePath, "-C", b.mountedRootFSPath)
}

func (b *RootFSBuilder) createAndMountEmptyRootFS() error {
	rootFSFile, err := ioutil.TempFile("", "oci-image-executor-root-fs-")
	if err != nil {
		return err
	}

	mountedRootFsPath, err := ioutil.TempDir("", "oci-image-executor-root-fs-mount-")
	if err != nil {
		return err
	}

	b.rootFSPath = rootFSFile.Name()
	b.mountedRootFSPath = mountedRootFsPath

	// fallocate is faster - see if this causes problems!
	// if err = RunCommandAndLogToStderr("dd", "if=/dev/zero", fmt.Sprintf("of=%s", b.rootFSPath), "bs=1M", "count=1500"); err != nil {
	//  	return err
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
	// Required?
	//if err = RunCommandAndLogToStderr("umount", b.rootFSPath); err != nil {
	//	return err
	//}

	if b.rootFSPath != "" {
		os.Remove(b.rootFSPath)
	}

	if b.mountedRootFSPath != "" {
		os.RemoveAll(b.mountedRootFSPath)
	}
}
