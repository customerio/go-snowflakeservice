//service layer to get metrics
package metrics_v3

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// func getMetricsAsync(){

// }

func getMetrics(ctx context.Context, params map[string][]string ) (Result, error) {
    result := Result{}

    //db
	db, dbErr := ctx.Value("db").(*sql.DB)

    if !dbErr {
        return result, errors.New("could not get database connection pool from context")
    }

    //build query here
    var query, q_err = buildQuery(params)
    if(q_err != nil){
        return result, q_err
    }

   // fmt.Println(query)

    //get metrics
    rows, queryErr := db.Query(query)
    if queryErr != nil {
        return result, queryErr
    }
    defer rows.Close()

    
    var allMetrics []MetricsModel
    for rows.Next() {

        var metricsModel MetricsModel
        rowErr := rows.Scan(&metricsModel.Campaign_ID, &metricsModel.Newsletter_id, &metricsModel.Created_At,
                            &metricsModel.Metric, &metricsModel.Metric_count)

        if rowErr != nil {
            return result, rowErr
        }
                
        allMetrics = append(allMetrics, metricsModel)
    }

    if rows.Err() != nil {
        return result, rows.Err()
    }

    result.WorkspaceID = params["workspace_id"][0];
    result.Metrics = &allMetrics;
    result.Next = ""

    return result, nil
}


func buildQuery(params map[string][]string) (string, error){
    //retrieve query parameters from the context

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
    var queryString =   "select d.campaign_id, d.newsletter_id, to_date(d.created_at), m.metric, count(metric) as metric_count" +"\n"+
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
    queryString += "group by d.campaign_id, d.newsletter_id, to_date(d.created_at), m.metric"

    return queryString, nil
}

func buildNewsletterQuery(params map[string][]string) (string,  error){
    var queryString =   "select d.campaign_id, d.newsletter_id, to_date(d.created_at), m.metric, count(metric) as metric_count" +"\n"+
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
    queryString += "group by d.campaign_id, d.newsletter_id, to_date(d.created_at), m.metric"

    return queryString, nil
}

func buildAll(params map[string][]string) (string,  error){
    var queryString =   "select d.campaign_id, d.newsletter_id, to_date(d.created_at), m.metric, count(metric) as metric_count" +"\n"+
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
    queryString += "group by d.campaign_id, d.newsletter_id, to_date(d.created_at), m.metric"

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