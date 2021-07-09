package database

import (
	"fmt"
	config "snowflakeservice/config"
)

func getConnectionString(env string) (string, error) {
	sfConfig, err := config.LoadSFConfig(env)
	if err != nil {
		return "", err
	}
	//"jsmith:mypassword@myaccount/mydb/testschema?warehouse=mywh"
	connString := fmt.Sprintf("%s:%s@%s/%s/%s?warehouse=%s",
		sfConfig.SF_Username, sfConfig.SF_Password, sfConfig.SF_Server,
		sfConfig.SF_DbName, sfConfig.SF_Schema, sfConfig.SF_Warehouse)
	return connString, nil

}
