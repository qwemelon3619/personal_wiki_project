package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// AnalyticsClient handles communication with analytics-service
type AnalyticsClient struct {
	baseURL string
	client  *http.Client
}

func NewAnalyticsClient(baseURL string) *AnalyticsClient {
	return &AnalyticsClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// RecordAPICall sends an API call record to analytics-service (async)
func (ac *AnalyticsClient) RecordAPICall(endpoint, userID string) {
	go func() {
		payload := map[string]string{
			"endpoint": endpoint,
			"userId":   userID,
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			log.Printf("[Analytics] Failed to marshal payload: %v", err)
			return
		}

		url := fmt.Sprintf("%s/api/analytics/record", ac.baseURL)
		resp, err := ac.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("[Analytics] Failed to record API call: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("[Analytics] API call recording failed with status: %d", resp.StatusCode)
		}
	}()
}
