package main

import (
	"encoding/json"
	"io/ioutil"
)

type ImageConfig struct {
	Cmd []string
	Env []string
}

func ImageConfigFromFile(filePath string) (ImageConfig, error) {
	fileContents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return ImageConfig{}, err
	}

	imageConfig := ImageConfig{}

	if err = json.Unmarshal(fileContents, &imageConfig); err != nil {
		return ImageConfig{}, err
	}

	return imageConfig, nil
}
