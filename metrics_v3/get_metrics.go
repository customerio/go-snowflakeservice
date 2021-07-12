//service layer to get metrics
package metrics_v3

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"snowflakeservice/database"
	"strconv"
	"strings"
	"time"
)

const (
	DELIVERIES_TABLE_NAME = "DELIVERIES"
	METRICS_TABLE_NAME    = "METRICS_UP_TO_20210603"
	PAGE_START_DEFAULT    = 1
	LIMIT_DEFAULT         = 10
)

func getMetrics(ctx context.Context, params map[string][]string) (Result, error) {
	result := Result{}

	dbs, err := ctx.Value("dbs").(*database.DBSessions)
	if !err {
		return result, errors.New("could not get database connection pool from context")
	}

	//db
	db := dbs.Sf_session
	if db == nil {
		return result, errors.New("could not get database connection pool from context")
	}

	//get page number and data size
	limit := LIMIT_DEFAULT //default
	offset := limit * (PAGE_START_DEFAULT - 1)

	size, ok := params["size"]
	if ok && len(size) == 1 {
		number, _ := strconv.Atoi(size[0])
		limit = number
	}

	page, ok := params["page"]
	if ok && len(page) == 1 {
		pg, _ := strconv.Atoi(page[0])
		offset = limit * (pg - 1)
	}

	//build query here
	var query, args, q_err = buildQuery(params)
	if q_err != nil {
		return result, q_err
	}

	query += fmt.Sprintf(" limit %d offset %d", limit, offset)

	//get data
	var allMetrics []MetricsModel

	queryErr := db.Select(&allMetrics, query, args...)

	if queryErr != nil {
		return result, queryErr
	}

	result.WorkspaceID = params["workspace_id"][0]
	result.Metrics = &allMetrics
	result.Next = ""

	return result, nil
}

func getMetricsAsync(ctx context.Context, params map[string][]string) error {

	r_err := generateReport(ctx, params)

	if r_err != nil {
		return r_err
	}
	return nil
}

//helpers
func generateReport(ctx context.Context, params map[string][]string) error {

	// reportChannel := make(chan *ReportModel)

	dbs, err := ctx.Value("dbs").(*database.DBSessions)
	if !err {
		return errors.New("could not get database connection pool from context")
	}

	//db
	db := dbs.Sf_session
	if db == nil {
		return errors.New("could not get Snowflake database connection pool from context")
	}

	//gcs
	gcsConf := dbs.GCS_Data
	if gcsConf == nil {
		return errors.New("could not get GCS details from context")
	}

	go func() {

		allMetrics := []MetricsModel{}

		//build query here
		var query, args, q_err = buildQuery(params)
		if q_err != nil {
			log.Fatal(q_err)
			return
		}

		queryErr := db.Select(&allMetrics, query, args...)
		if queryErr != nil {
			log.Fatal(queryErr)
			return
		}

		//generate report using data
		var filename = params["workspace_id"][0] + "_metrics_local_" + time.Now().Format(time.RFC3339Nano) + ".csv"
		_, f_err := createCSVFile(allMetrics, filename)
		if f_err != nil {
			log.Fatal(f_err)
			return
		}

		//save report
		nameOfCloudObject := params["workspace_id"][0] + "_metrics_cloud_" + time.Now().Format(time.RFC3339Nano) + ".csv"
		_err := database.UploadPath(filename, nameOfCloudObject)
		if _err != nil {
			log.Fatal(_err)
			return
		}

		//return signed url -- expires 4hrs from creation time
		downloadURL, err := database.SignedURL(nameOfCloudObject, time.Duration(14400000000000), *gcsConf)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Download URL : " + downloadURL)

	}()

	return nil
}

func createCSVFile(data []MetricsModel, filepath string) (*os.File, error) {
	file, err := os.Create(filepath)
	if err != nil {
		return file, err
	}
	defer file.Close()

	dataWriter := csv.NewWriter(file)

	for _, mm := range data {
		var row []string
		if mm.Campaign_ID != nil {
			row = append(row, *mm.Campaign_ID)
		} else {
			row = append(row, "")
		}
		if mm.Newsletter_ID != nil {
			row = append(row, *mm.Newsletter_ID)
		} else {
			row = append(row, "")
		}
		row = append(row, mm.Creates_At.String())
		row = append(row, mm.Metric)
		row = append(row, strconv.Itoa(mm.Metric_Count))
		dataWriter.Write(row)
	}

	dataWriter.Flush()

	return file, nil
}

func buildQuery(params map[string][]string) (string, []interface{}, error) {
	//if campaign
	isCampaign, c_ok := params["isCampaign"]
	if c_ok && len(isCampaign) == 1 && isCampaign[0] == "true" {
		return buildCampaignQuery(params)
	}

	isNewsletter, n_ok := params["isNewsletter"]
	if n_ok && len(isNewsletter) == 1 && isNewsletter[0] == "true" {
		return buildNewsletterQuery(params)
	}

	//add more object filters here

	//all inclusive
	return buildAll(params)
}

func buildCampaignQuery(params map[string][]string) (string, []interface{}, error) {
	var args []interface{}

	var queryString = "select d.campaign_id, d.newsletter_id, to_date(d.created_at) as created_at, m.metric, count(metric) as metric_count " +
		"from DELIVERIES d " +
		"join METRICS_UP_TO_20210603 m " +
		"on d.delivery_id = m.delivery_id " +
		"where "

	//build common params
	queryString, args, err := buildWithCommonParams(params, queryString, args)
	if err != nil {
		return "", args, err
	}

	//after all query options
	queryString += "and d.campaign_id is not null "
	queryString += "group by d.campaign_id, d.newsletter_id, to_date(d.created_at), m.metric " +
		"order by d.campaign_id, to_date(d.created_at) "

	return queryString, args, nil
}

func buildNewsletterQuery(params map[string][]string) (string, []interface{}, error) {
	var args []interface{}

	var queryString = "select d.campaign_id, d.newsletter_id, to_date(d.created_at) as created_at, m.metric, count(metric) as metric_count " +
		"from ? d " +
		"join ? m " +
		"on d.delivery_id = m.delivery_id " +
		"where "

	args = append(args, DELIVERIES_TABLE_NAME)
	args = append(args, METRICS_TABLE_NAME)

	//build common params
	queryString, args, err := buildWithCommonParams(params, queryString, args)
	if err != nil {
		return "", args, err
	}

	//after all query options
	queryString += "and d.newsletter_id is not null " +
		"group by d.campaign_id, d.newsletter_id, to_date(d.created_at), m.metric " +
		"order by d.newsletter_id, to_date(d.created_at) "

	return queryString, args, nil
}

func buildAll(params map[string][]string) (string, []interface{}, error) {
	var args []interface{}

	var queryString = "select d.campaign_id, d.newsletter_id, to_date(d.created_at) as created_at, m.metric, count(metric) as metric_count" +
		"from ? d " +
		"join ? m " +
		"on d.delivery_id = m.delivery_id " +
		"where "

	args = append(args, DELIVERIES_TABLE_NAME)
	args = append(args, METRICS_TABLE_NAME)

	//build common params
	queryString, args, err := buildWithCommonParams(params, queryString, args)
	if err != nil {
		return "", args, err
	}

	//after all query options
	queryString += "group by d.campaign_id, d.newsletter_id, to_date(d.created_at), m.metric " +
		"order by d.campaign_id, to_date(d.created_at) "

	return queryString, args, nil
}

func buildWithCommonParams(params map[string][]string, query string, args []interface{}) (string, []interface{}, error) {

	//workspace_id
	workspace_id, w_ok := params["workspace_id"]

	if !w_ok || len(workspace_id) != 1 {
		return "", nil, errors.New("only one value of Workspace_id is required")
	} else {
		query += "d.workspace_id = ? "
	}
	args = append(args, workspace_id[0])

	//start date and end
	start_date, s_ok := params["start"]
	if s_ok && len(start_date) != 1 {
		return "", nil, errors.New("only one value of start is required")

	} else if s_ok && len(start_date) == 1 {

		end_date, e_ok := params["end"]
		if !e_ok || len(end_date) != 1 {
			return "", nil, errors.New("only one value of end is required if there's a start date")
		} else {
			query += "and to_date(d.created_at) between ? and ? "
			args = append(args, start_date[0])
			args = append(args, end_date[0])
		}

	}

	//metric
	metric, m_ok := params["metric"]
	if m_ok {
		query += "and m.metric in (?" + strings.Repeat(",?", len(metric)-1) + ") "
		for data := range metric {
			args = append(args, metric[data])
		}
	}

	//multiple dates
	dates, d_ok := params["dates"]
	if d_ok {
		query += "and to_date(d.created_at) in (?" + strings.Repeat(",?", len(dates)-1) + ") "
		for data := range dates {
			args = append(args, dates[data])
		}
	}

	//delivery_type
	d_types, dt_ok := params["dtype"]
	if dt_ok {
		query += "and d.delivery_type in (?" + strings.Repeat(",?", len(d_types)-1) + ")"
		for data := range d_types {
			args = append(args, d_types[data])
		}
	}

	//include other query options here

	return query, args, nil
}
