package database

import (
	config "snowflakeservice/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/snowflakedb/gosnowflake"
)

type DBSessions struct {
	GCS_Data   *config.GCSConfig
	Sf_session *sqlx.DB
}

func InitDB(env string) (*DBSessions, error) {
	var dbs *DBSessions

	//get snowflake session
	sfConnString, err := getConnectionString(env)
	if err != nil {
		return nil, err
	}

	db, err := sqlx.Open("snowflake", sfConnString)
	if err != nil {
		return nil, err
	}

	dbs = &DBSessions{}
	dbs.Sf_session = db

	//get gcs info
	gcsConf, err := getConfig(env)
	if err != nil {
		return nil, err
	}

	dbs.GCS_Data = gcsConf

	return dbs, nil
}
