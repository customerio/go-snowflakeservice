package database

import (
	"fmt"
	config "snowflakeservice/config"
)

func getConnectionString() string{
	server := config.SF_SERVER
	userName := config.SF_USERNAME
	password := config.SF_PASSWORD
	dbName := config.SF_DBNAME
	schema := config.SF_SCHEMA
	warehouseName := config.SF_WAREHOUSE

	//"jsmith:mypassword@myaccount/mydb/testschema?warehouse=mywh"
	return fmt.Sprintf("%s:%s@%s/%s/%s?warehouse=%s",
	 userName, password, server, dbName, schema, warehouseName);

}
