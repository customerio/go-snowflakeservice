package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	sf "snowflakeservice/database"
	index "snowflakeservice/home"
	metricsApi "snowflakeservice/metrics_v3"

	"github.com/gorilla/mux"
	_ "github.com/snowflakedb/gosnowflake"
)

//inject db connection into reuests
func serveDB(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := context.WithValue(r.Context(), "db", db)

        next.ServeHTTP(w, r.WithContext(ctx))
    }
}



func main() {
	db, err := sql.Open("snowflake", sf.GetConnectionString())
    if err != nil {
		
        log.Fatal(err)
    }

    router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", serveDB(db, index.HomeLink))
	router.HandleFunc("/api/metrics/search", serveDB(db, metricsApi.SearchMetrics))
	log.Fatal(http.ListenAndServe(":8089", router))
}