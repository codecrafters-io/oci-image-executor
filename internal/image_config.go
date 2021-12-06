package internal

import (
	"encoding/json"
	"io/ioutil"

	"github.com/tidwall/gjson"
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

	imageConfig := ImageConfig{
		Cmd: []string{},
		Env: []string{},
	}

	for _, cmdResult := range gjson.Get(string(fileContents), "config.Cmd").Array() {
		imageConfig.Cmd = append(imageConfig.Cmd, cmdResult.String())
	}

	for _, envResult := range gjson.Get(string(fileContents), "config.Env").Array() {
		imageConfig.Env = append(imageConfig.Env, envResult.String())
	}

	if err = json.Unmarshal(fileContents, &imageConfig); err != nil {
		return ImageConfig{}, err
	}

	return imageConfig, nil
}
