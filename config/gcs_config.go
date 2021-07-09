package config

import (
	"fmt"
	"strings"

	"github.com/tkanos/gonfig"
)

type GCSConfig struct {
	Project_Id   string
	Private_Key  string
	Client_Email string
	Client_ID    string
	Auth_URI     string
	Token_URI    string
}

func LoadGCSConfig(env string) (GCSConfig, error) {
	var gcsConfig GCSConfig

	currentEnv := "dev"
	if env != "" {
		currentEnv = strings.ToLower(env)
	}

	fileName := fmt.Sprintf("./gcs_%s.json", currentEnv)

	err := gonfig.GetConf(fileName, &gcsConfig)
	if err != nil {
		return gcsConfig, err
	}
	return gcsConfig, nil
}
