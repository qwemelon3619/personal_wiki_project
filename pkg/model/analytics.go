package model

import "time"

type APICallLog struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	Endpoint  string    `bson:"endpoint" json:"endpoint"`
	UserID    string    `bson:"userId" json:"userId"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}

type AnalyticsSummary struct {
	TotalCalls   int64           `json:"totalCalls"`
	UniqueUsers  int64           `json:"uniqueUsers"`
	TopEndpoints []EndpointCount `json:"topEndpoints"`
}

type EndpointStats struct {
	Endpoint     string    `json:"endpoint"`
	TotalCalls   int64     `json:"totalCalls"`
	UniqueUsers  int64     `json:"uniqueUsers"`
	LastAccessed time.Time `json:"lastAccessed"`
}

type UserStats struct {
	UserID       string          `json:"userId"`
	TotalCalls   int64           `json:"totalCalls"`
	TopEndpoints []EndpointCount `json:"topEndpoints"`
}

type EndpointCount struct {
	Endpoint string `json:"endpoint"`
	Count    int64  `json:"count"`
}
