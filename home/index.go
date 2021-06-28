package home

import (
	"fmt"
	"net/http"
)

//landing page
func HomeLink(writer http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(writer, "Metrics Data Application with Snowflake!")
}