//service layer to get metrics
package metrics_v3

import (
	"context"
	//	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	//  databaseConfig "snowflakeservice/database"

	"github.com/jmoiron/sqlx"
)

func getMetrics(ctx context.Context, params map[string][]string ) (Result, error) {
    result := Result{}

    //db
	db, err := ctx.Value("db").(*sqlx.DB)
    if !err {
        return result, errors.New("could not get database connection pool from context")
    }

    //get page number and data size
    limit := 10; //default
    offset := 0;

    size, ok := params["size"]
    if ok && len(size) == 1{
        number, _ := strconv.Atoi(size[0])
        limit = number 
    }

    page, ok := params["page"]
    if(ok && len(page) == 1){
        pg, _ := strconv.Atoi(page[0])
        offset = limit * (pg - 1)
    }


    //build query here
    var query, q_err = buildQuery(params)
    if(q_err != nil){
        return result, q_err
    }
    query += fmt.Sprintf(" limit %d offset %d", limit, offset)


    //get data
    var allMetrics []MetricsModel //= []MetricsModel{}
    queryErr := db.Select(&allMetrics,  query)

    if queryErr != nil {
        return result, queryErr
    }

    result.WorkspaceID = params["workspace_id"][0];
    result.Metrics = &allMetrics;
    result.Next = ""

    return result, nil
}

func getMetricsAsync(ctx context.Context, params map[string][]string) (r *ReportModel, e error){
    
    reportModel, r_err := generateReport(ctx, params)
    
    if e, ok := <-r_err ; ok{
        return r, e
    }
  
    r = <-reportModel
    return r,nil
 }


 //helpers
func generateReport(ctx context.Context, params map[string][]string) (chan *ReportModel, chan error){

    reportChannel := make(chan *ReportModel)
    errorChannel := make(chan error)

    go func(){
        // defer close(reportChannel)
        // defer close(errorChannel)
        
        //db
        db, err := ctx.Value("db").(*sqlx.DB)
        if !err {
            errorChannel <- errors.New("could not get database connection pool from context")
            return 
        }

        allMetrics := []MetricsModel{}

        //build query here
        var query, q_err = buildQuery(params)
        if(q_err != nil){
            errorChannel <- q_err
            return
        }

        queryErr := db.Select(&allMetrics,  query)

        if queryErr != nil {
            errorChannel <- queryErr
            return
        }

        //generate report using data 
        var filePath = "metrics.csv"
        file, f_err := createCSVFile(allMetrics, filePath)
        if(f_err != nil){
            errorChannel <- f_err
            return
        }

        //save report
        reportModel, _err := uploadFileToS3(ctx, file)
        if(_err != nil){
            errorChannel <- _err
            return
        }

        reportChannel <- reportModel

    }()

    defer close(reportChannel)
    defer close(errorChannel)

    return reportChannel, errorChannel    
}

func createCSVFile(data []MetricsModel, filepath string) (*os.File, error){
    file, err := os.Create(filepath)
    
    defer file.Close()
    if err != nil {
        return file, err
    }

    dataWriter := csv.NewWriter(file)
    
    for _, mm := range data {
        var row []string
        row = append(row, *mm.Campaign_ID)
        row = append(row, *mm.Newsletter_ID)
        row = append(row, mm.Creates_At.String())
        row = append(row, *&mm.Metric)
        row = append(row, strconv.Itoa(mm.Metric_Count))
        dataWriter.Write(row)
    }

    dataWriter.Flush()
  
    return file, nil
}

func uploadFileToS3(ctx context.Context, file *os.File) (*ReportModel, error){

    reportModel := ReportModel{URL: "http://test.test"}
    return &reportModel, nil
}

func buildQuery(params map[string][]string) (string, error){
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

func buildCampaignQuery(params map[string][]string) (string,  error){
    var queryString =   "select d.campaign_id, d.newsletter_id, to_date(d.created_at) as created_at, m.metric, count(metric) as metric_count" +"\n"+
                        "from DELIVERIES d " + "\n" +
                        "join metrics_up_to_20210603 m " + "\n" +
                        "on d.delivery_id = m.delivery_id " + "\n" +
                        "where " 

    //build common params
    subquery, err := buildWithCommonParams(params)
    if err != nil {
       return "", err
    }

    queryString += subquery

    //after all query options
    queryString += "and d.campaign_id is not null "
    queryString += "group by d.campaign_id, d.newsletter_id, to_date(d.created_at), m.metric " + "\n" +
                    "order by d.campaign_id, to_date(d.created_at) "

    return queryString, nil
}

func buildNewsletterQuery(params map[string][]string) (string,  error){
    var queryString =   "select d.campaign_id, d.newsletter_id, to_date(d.created_at) as created_at, m.metric, count(metric) as metric_count" +"\n"+
                        "from DELIVERIES d " + "\n" +
                        "join metrics_up_to_20210603 m " + "\n" +
                        "on d.delivery_id = m.delivery_id " + "\n" +
                        "where " 

     //build common params
    subquery, err := buildWithCommonParams(params)
    if err != nil {
       return "", err
    }

    queryString += subquery

    //after all query options
    queryString += "and d.newsletter_id is not null "
    queryString += "group by d.campaign_id, d.newsletter_id, to_date(d.created_at), m.metric " + "\n" +
                    "order by d.newsletter_id, to_date(d.created_at) "

    return queryString, nil
}

func buildAll(params map[string][]string) (string,  error){
    var queryString =   "select d.campaign_id, d.newsletter_id, to_date(d.created_at) as created_at, m.metric, count(metric) as metric_count" +"\n"+
                        "from DELIVERIES d " + "\n" +
                        "join metrics_up_to_20210603 m " + "\n" +
                        "on d.delivery_id = m.delivery_id " + "\n" +
                        "where " 

     //build common params
    subquery, err := buildWithCommonParams(params)
    if err != nil {
       return "", err
    }

    queryString += subquery

    //after all query options
    queryString += "group by d.campaign_id, d.newsletter_id, to_date(d.created_at), m.metric " + "\n" +
                    "order by d.campaign_id, to_date(d.created_at) "

    return queryString, nil
}

func buildWithCommonParams (params map[string][]string) (string,  error){
    queryString := ""

     //workspace_id
    workspace_id, w_ok := params["workspace_id"]

    if !w_ok || len(workspace_id) != 1  {
        return "" , errors.New("only one value of Workspace_id is required")
    }else {
        queryString += fmt.Sprintf("d.workspace_id = %s ", workspace_id[0])
    }


    //start date and end
    start_date, s_ok := params["start"]
    if s_ok && len(start_date) != 1  {
    
        return "" , errors.New("only one value of start is required")

    }else if s_ok && len(start_date) == 1 {

        end_date , e_ok := params["end"]
        if !e_ok || len(end_date) != 1{
            return "" , errors.New("only one value of end is required if there's a start date")
        } else{
            queryString += fmt.Sprintf("and to_date(d.created_at) between '%s' and '%s' ", start_date[0], end_date[0])
        }
       
    }

    //metric
    metric, m_ok := params["metric"]
    if(m_ok){
        val := "'"
        val += strings.Replace(strings.Trim(fmt.Sprint(metric), "[]"), " ", "','", -1)
        val += "'"
        
        queryString += fmt.Sprintf("and m.metric in (%s) ", val)
    }


    //multiple dates
    dates, d_ok := params["dates"]
    if(d_ok){
        val := "'"
        val += strings.Replace(strings.Trim(fmt.Sprint(dates), "[]"), " ", "','", -1)
        val += "'"
        
        queryString += fmt.Sprintf("and to_date(d.created_at) in (%s) ", val)
    }


    //delivery_type
    d_types, dt_ok := params["dtype"]
    if(dt_ok){
        val := "'"
        val += strings.Replace(strings.Trim(fmt.Sprint(d_types), "[]"), " ", "','", -1)
        val += "'"
        queryString += fmt.Sprintf("and d.delivery_type in (%s) ", val)
    }

    //include other query options here

    return queryString, nil
}