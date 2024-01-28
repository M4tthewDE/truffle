package config

import (
	"encoding/json"
	"io"
	"os"
)

var (
	Conf Config
)

type Config struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Url          string `json:"url"`
}

func LoadConfig() error {
	jsonFile, err := os.Open("config.json")
	if err != nil {
		return err
	}

	defer jsonFile.Close()

	jsonBytes, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	var conf Config
	err = json.Unmarshal(jsonBytes, &conf)
	if err != nil {
		return err
	}

	Conf = conf

	return nil
}
