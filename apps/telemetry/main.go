package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type TelemetryData struct {
	DeviceID  string         `json:"device_id"`
	Timestamp time.Time      `json:"timestamp"`
	Metrics   map[string]any `json:"metrics"`
}

type TelemetryHistory struct {
	DeviceID string          `json:"device_id"`
	Data     []TelemetryData `json:"data"`
}

func main() {
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Get latest telemetry for device
	router.GET("/telemetry/:deviceId/latest", func(c *gin.Context) {
		deviceID := c.Param("deviceId")

		// Check if device ID is valid (1, 2, 3, or 4)
		if !isValidDeviceID(deviceID) {
			c.JSON(http.StatusNotFound, gin.H{"error": "device with id not found"})
			return
		}

		// Generate mock telemetry data
		data := generateLatestTelemetry(deviceID)
		c.JSON(http.StatusOK, data)
	})

	// Get telemetry history for device
	router.GET("/telemetry/:deviceId/history", func(c *gin.Context) {
		deviceID := c.Param("deviceId")

		// Check if device ID is valid (1, 2, 3, or 4)
		if !isValidDeviceID(deviceID) {
			c.JSON(http.StatusNotFound, gin.H{"error": "device with id not found"})
			return
		}

		// Get query parameters for date range
		from := c.Query("from")
		to := c.Query("to")

		// Generate mock telemetry history
		history := generateTelemetryHistory(deviceID, from, to)
		c.JSON(http.StatusOK, history)
	})

	// Start server
	log.Println("Telemetry API starting on :8082")
	if err := router.Run(":8082"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func isValidDeviceID(deviceID string) bool {
	// Convert to int to check if it's 1, 2, 3, or 4
	id, err := strconv.Atoi(deviceID)
	if err != nil {
		return false
	}
	return id >= 1 && id <= 4
}

func generateLatestTelemetry(deviceID string) TelemetryData {
	// Generate different metrics based on device ID
	metrics := make(map[string]any)

	switch deviceID {
	case "1":
		metrics["power"] = 23.5
		metrics["voltage"] = 220.1
		metrics["current"] = 0.11
	case "2":
		metrics["temperature"] = 22.3
		metrics["humidity"] = 45.2
		metrics["pressure"] = 1013.25
	case "3":
		metrics["power"] = 15.7
		metrics["voltage"] = 220.0
		metrics["current"] = 0.07
	case "4":
		metrics["temperature"] = 24.1
		metrics["humidity"] = 48.7
		metrics["pressure"] = 1012.8
	}

	return TelemetryData{
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		Metrics:   metrics,
	}
}

func generateTelemetryHistory(deviceID, from, to string) TelemetryHistory {
	// Generate mock history data
	history := TelemetryHistory{
		DeviceID: deviceID,
		Data:     make([]TelemetryData, 0),
	}

	// Generate 5 data points with decreasing timestamps
	for i := range 5 {
		timestamp := time.Now().Add(-time.Duration(i*5) * time.Minute)

		metrics := make(map[string]any)

		switch deviceID {
		case "1":
			metrics["power"] = 20.1 + float64(i)*0.8
			metrics["voltage"] = 220.0 + float64(i)*0.1
			metrics["current"] = 0.09 + float64(i)*0.005
		case "2":
			metrics["temperature"] = 21.5 + float64(i)*0.3
			metrics["humidity"] = 44.0 + float64(i)*0.8
			metrics["pressure"] = 1013.0 + float64(i)*0.1
		case "3":
			metrics["power"] = 14.2 + float64(i)*0.5
			metrics["voltage"] = 219.8 + float64(i)*0.1
			metrics["current"] = 0.06 + float64(i)*0.003
		case "4":
			metrics["temperature"] = 23.8 + float64(i)*0.2
			metrics["humidity"] = 47.5 + float64(i)*0.4
			metrics["pressure"] = 1012.5 + float64(i)*0.1
		}

		data := TelemetryData{
			DeviceID:  deviceID,
			Timestamp: timestamp,
			Metrics:   metrics,
		}
		history.Data = append(history.Data, data)
	}

	return history
}
