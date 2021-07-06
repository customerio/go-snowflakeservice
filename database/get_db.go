package database

import (
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jmoiron/sqlx"

	_ "github.com/snowflakedb/gosnowflake"
)

type DBSessions struct {
	S3_session *session.Session
	Sf_session *sqlx.DB
}

func InitDB() *DBSessions {
	dbs := DBSessions{}

	//get snowflake session
	db, err := sqlx.Open("snowflake", getConnectionString())
    if err != nil {
        log.Fatal(err)
    }
	dbs.Sf_session = db

	awsSession, err := newSession()
	if( err != nil){
		log.Fatal(err)
	}
	dbs.S3_session = awsSession

	return &dbs;
}