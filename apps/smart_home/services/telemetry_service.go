package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// TelemetryService handles fetching telemetry data from external API
type TelemetryService struct {
	BaseURL    string
	HTTPClient *http.Client
}

// TelemetryData represents the response from the telemetry API
type TelemetryData struct {
	DeviceID  string         `json:"device_id"`
	Timestamp time.Time      `json:"timestamp"`
	Metrics   map[string]any `json:"metrics"`
}

// TelemetryHistory represents the history response from the telemetry API
type TelemetryHistory struct {
	DeviceID string          `json:"device_id"`
	Data     []TelemetryData `json:"data"`
}

// NewTelemetryService creates a new telemetry service
func NewTelemetryService(baseURL string) *TelemetryService {
	return &TelemetryService{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetLatestTelemetry fetches latest telemetry data for a specific device
func (s *TelemetryService) GetLatestTelemetry(deviceID string) (*TelemetryData, error) {
	url := fmt.Sprintf("%s/telemetry/%s/latest", s.BaseURL, deviceID)

	resp, err := s.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching telemetry data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("device with id not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var telemetryResp TelemetryData
	if err := json.NewDecoder(resp.Body).Decode(&telemetryResp); err != nil {
		return nil, fmt.Errorf("error decoding telemetry response: %w", err)
	}

	return &telemetryResp, nil
}

// GetTelemetryHistory fetches telemetry history for a specific device
func (s *TelemetryService) GetTelemetryHistory(deviceID string, from, to string) (*TelemetryHistory, error) {
	baseURL := fmt.Sprintf("%s/telemetry/%s/history", s.BaseURL, deviceID)

	// Build URL with query parameters
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %w", err)
	}

	// Add query parameters if provided
	params := u.Query()
	if from != "" {
		params.Set("from", from)
	}
	if to != "" {
		params.Set("to", to)
	}
	u.RawQuery = params.Encode()

	url := u.String()

	resp, err := s.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching telemetry history: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("device with id not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var historyResp TelemetryHistory
	if err := json.NewDecoder(resp.Body).Decode(&historyResp); err != nil {
		return nil, fmt.Errorf("error decoding telemetry history response: %w", err)
	}

	return &historyResp, nil
}
