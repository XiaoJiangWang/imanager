package util

import "time"

type ErrorResponse struct {
	ErrorCode    int    `json:"ErrorCode"`
	ErrorMessage string `json:"ErrorMessage"`
}

type BaseModel struct {
	CreateTimestamp time.Time  `json:"create_timestamp" orm:"column(create_timestamp)"`
	UpdateTimestamp time.Time  `json:"update_timestamp" orm:"column(update_timestamp)"`
}