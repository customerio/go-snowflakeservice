package database

import "fmt"

func GetConnectionString() string{
	server := "ciodevtest.us-central1.gcp"
	userName := "adedamola"
	password := "<password here>"
	dbName := "CIO_DELIVERIES"
	schema := "PUBLIC"
	warehouseName := "COMPUTE_WH"

	//"jsmith:mypassword@myaccount/mydb/testschema?warehouse=mywh"
	return fmt.Sprintf("%s:%s@%s/%s/%s?warehouse=%s",
	 userName, password, server, dbName, schema, warehouseName);

}
