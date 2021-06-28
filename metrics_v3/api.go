//http layer to get metrics
package metrics_v3

import (
	"encoding/json"
	"net/http"
)

func SearchMetrics(writer http.ResponseWriter, request *http.Request){
	params := request.URL.Query();

	data, err := getMetrics(request.Context(), params)
		if err != nil {
			returnErrorAsJson(writer, 500, err.Error())
			return
		}
	returnResultAsJson(writer, 200, data)
}

func returnErrorAsJson(writer http.ResponseWriter, code int, message string){
	returnResultAsJson(writer, code, map[string]string{"error":message})
}

func returnResultAsJson(writer http.ResponseWriter, code int, data interface{}){
	response,_ := json.Marshal(data)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(code)
	writer.Write(response)
}