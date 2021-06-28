package metrics_v3

import (
	"time"
)

type MetricsModel struct{
	Campaign_ID *string "json:campaign_id"
	Newsletter_id *string "json:newsletter_id"
	Created_At time.Time "json:created_at"
	Metric string "json:metric"
	Metric_count int "json:value" 
}

type Result struct{
	WorkspaceID string "json:workspace_id"	
	Metrics *[]MetricsModel "json:metrics"
	Next string "json:next"

}

