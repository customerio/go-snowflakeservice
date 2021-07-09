package main

import (
	"context"
	"log"
	"net/http"
	database "snowflakeservice/database"
	index "snowflakeservice/home"
	metricsApi "snowflakeservice/metrics_v3"

	"github.com/gorilla/mux"
)

//inject db connection into reuests
func loadDB(dbs *database.DBSessions, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "dbs", dbs)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func main() {
	env := "dev" //load environment

	dbs, err := database.InitDB(env)
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", loadDB(nil, index.HomeLink))
	router.HandleFunc("/api/metrics/search", loadDB(dbs, metricsApi.SearchMetrics))
	router.HandleFunc("/api/metrics/report", loadDB(dbs, metricsApi.GenerateReport))
	log.Fatal(http.ListenAndServe(":8089", router))
}
