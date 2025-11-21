package models

import "time"

type Item struct {
	ID    int       `json:"id,omitempty"`
	Price int       `json:"price"`
	Name  string    `json:"name"`
	Type  string    `json:"type"`
	Date  time.Time `json:"date,omitempty"`
}

type AnalyticsRequest struct {
	Type string
	Date *time.Time
	From *time.Time
	To   *time.Time
}

type AnalyticsResponse struct {
	Type         string  `json:"type"`
	Sum          int     `json:"sum,omitempty"`
	Avg          float64 `json:"avg,omitempty"`
	Count        int     `json:"count,omitempty"`
	Median       float64 `json:"median,omitempty"`
	Percentile90 float64 `json:"percentile_90,omitempty"`
}
