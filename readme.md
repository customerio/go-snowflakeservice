## Data Application with Snowflake
This service provides access to query data from the Snowflakes Storage which consists of large amounts of data

## Endpoints
There are two endpoints in this application
- GET "/" : this is the landing page
- GET "/api/metrics/search?<queryOptions>" : this is the main endpoint used to search the data store based on certain parameters
- GET "/api/metrics/report?<queryOptions>" : this is the an endpoint used to search the data store based on certain parameters using an async feature. It generates a report in the background and returns the download url when finished

### Query Options
The available query options are listed below

- workspace_id :  this is a **<required>** parameter used to return metrics for a specific client. It takes a numeric value
- start : **<not_required>** This specifies the start date in 'yyyy-mm-dd" format
- end : **<not_required>** however it must be present if the start parameter is present. Also uses the 'yyyy-mm-dd' format
- isCampaign : **<not_required>** This parameter indicates if only campaigns should be included in the search. Values = true, false
- isNewsletter : **<not_required>** This parameter indicates if only newsletterd should be included in the search. Values = true, false
                 if neither isCampaign nor isNewsletter options are present, the serch is all inclusive
- dates : **<not_required**> This parameter allows you to specify specific dats to search for. It can take in multiple values. E.g
"dates=2020-05-31&dates=2019-05-31&dates=2018-05-31"
- dtypes : **<not_required>** This parameter allows you to specify delivery_types to search for. It can take in multiple values. E.g 
"dtypes=email&dtypes=push"
- metric : **<not_required>** This parameter allows you to specify metrics to seach for. It can also take in multiple values. E.g
"metric=clicked&metric=sent&metric=unsubscribed