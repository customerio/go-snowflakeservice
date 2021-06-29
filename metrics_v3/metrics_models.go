package metrics_v3

import (
	"time"
)

type MetricsModel struct{
	Campaign_ID *string `json:"campaign_id" db:"CAMPAIGN_ID"`
	Newsletter_ID *string `json:"newsletter_id" db:"NEWSLETTER_ID"`
	Creates_At time.Time `json:"created_at" db:"CREATED_AT"`
	Metric string `json:"metric" db:"METRIC"`
	Metric_Count int `json:"value" db:"METRIC_COUNT"` 
}

type Result struct{
	WorkspaceID string `json:"workspace_id"`	
	Metrics *[]MetricsModel `json:"metrics"`
	Next string `json:"next"`

}

