package config

import (
	"fmt"
	"strings"

	"github.com/tkanos/gonfig"
)

type SFConfig struct {
	SF_Server    string `json:"SF_SERVER"`
	SF_Username  string `json:"SF_USERNAME"`
	SF_Password  string `json:"SF_PASSWORD"`
	SF_DbName    string `json:"SF_DBNAME"`
	SF_Schema    string `json:"SF_SCHEMA"`
	SF_Warehouse string `json:"SF_WAREHOUSE"`
}

func LoadSFConfig(env string) (SFConfig, error) {
	var sfConfig SFConfig

	currentEnv := "dev"
	if env != "" {
		currentEnv = strings.ToLower(env)
	}

	fileName := fmt.Sprintf("./sf_%s.json", currentEnv)

	err := gonfig.GetConf(fileName, &sfConfig)
	if err != nil {
		return sfConfig, err
	}
	return sfConfig, nil
}
