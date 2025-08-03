package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"smarthome/services"

	"github.com/gin-gonic/gin"
)

// SensorHandler handles sensor-related requests
type SensorHandler struct {
	DeviceManagementURL string
	TemperatureService  *services.TemperatureService
	TelemetryService    *services.TelemetryService
	HTTPClient          *http.Client
}

// NewSensorHandler creates a new SensorHandler
func NewSensorHandler(deviceManagementURL string, temperatureService *services.TemperatureService, telemetryService *services.TelemetryService) *SensorHandler {
	return &SensorHandler{
		DeviceManagementURL: deviceManagementURL,
		TemperatureService:  temperatureService,
		TelemetryService:    telemetryService,
		HTTPClient:          &http.Client{},
	}
}

// RegisterRoutes registers the sensor routes
func (h *SensorHandler) RegisterRoutes(router *gin.RouterGroup) {
	sensors := router.Group("/sensors")
	{
		sensors.GET("", h.GetSensors)
		sensors.GET("/:id", h.GetSensorByID)
		sensors.POST("", h.CreateSensor)
		sensors.PUT("/:id", h.UpdateSensor)
		sensors.DELETE("/:id", h.DeleteSensor)
		sensors.PATCH("/:id/value", h.UpdateSensorValue)
		sensors.GET("/temperature/:location", h.GetTemperatureByLocation)
	}

	// Telemetry routes
	telemetry := router.Group("/telemetry")
	{
		telemetry.GET("/:deviceId/latest", h.GetLatestTelemetry)
		telemetry.GET("/:deviceId/history", h.GetTelemetryHistory)
	}
}

// proxyRequest forwards the request to device-management service
func (h *SensorHandler) proxyRequest(c *gin.Context, method, path string, body io.Reader) {
	url := fmt.Sprintf("%s%s", h.DeviceManagementURL, path)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// Copy headers from original request
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Set content type for requests with body
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to proxy request: %v", err)})
		return
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	// Forward response status and body
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// GetSensors handles GET /api/v1/sensors
func (h *SensorHandler) GetSensors(c *gin.Context) {
	h.proxyRequest(c, "GET", "/api/v1/sensors", nil)
}

// GetSensorByID handles GET /api/v1/sensors/:id
func (h *SensorHandler) GetSensorByID(c *gin.Context) {
	id := c.Param("id")
	h.proxyRequest(c, "GET", fmt.Sprintf("/api/v1/sensors/%s", id), nil)
}

// GetTemperatureByLocation handles GET /api/v1/sensors/temperature/:location
func (h *SensorHandler) GetTemperatureByLocation(c *gin.Context) {
	location := c.Param("location")
	if location == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Location is required"})
		return
	}

	// Fetch temperature data from the external API
	tempData, err := h.TemperatureService.GetTemperature(location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to fetch temperature data: %v", err),
		})
		return
	}

	// Return the temperature data
	c.JSON(http.StatusOK, gin.H{
		"location":    tempData.Location,
		"value":       tempData.Value,
		"unit":        tempData.Unit,
		"status":      tempData.Status,
		"timestamp":   tempData.Timestamp,
		"description": tempData.Description,
	})
}

// CreateSensor handles POST /api/v1/sensors
func (h *SensorHandler) CreateSensor(c *gin.Context) {
	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	h.proxyRequest(c, "POST", "/api/v1/sensors", bytes.NewReader(body))
}

// UpdateSensor handles PUT /api/v1/sensors/:id
func (h *SensorHandler) UpdateSensor(c *gin.Context) {
	id := c.Param("id")

	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	h.proxyRequest(c, "PUT", fmt.Sprintf("/api/v1/sensors/%s", id), bytes.NewReader(body))
}

// DeleteSensor handles DELETE /api/v1/sensors/:id
func (h *SensorHandler) DeleteSensor(c *gin.Context) {
	id := c.Param("id")
	h.proxyRequest(c, "DELETE", fmt.Sprintf("/api/v1/sensors/%s", id), nil)
}

// UpdateSensorValue handles PATCH /api/v1/sensors/:id/value
func (h *SensorHandler) UpdateSensorValue(c *gin.Context) {
	id := c.Param("id")

	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	h.proxyRequest(c, "PATCH", fmt.Sprintf("/api/v1/sensors/%s/value", id), bytes.NewReader(body))
}

// GetLatestTelemetry handles GET /api/v1/telemetry/:deviceId/latest
func (h *SensorHandler) GetLatestTelemetry(c *gin.Context) {
	deviceID := c.Param("deviceId")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device ID is required"})
		return
	}

	// Fetch telemetry data from the external API
	telemetryData, err := h.TelemetryService.GetLatestTelemetry(deviceID)
	if err != nil {
		if err.Error() == "device with id not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "device with id not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to fetch telemetry data: %v", err),
		})
		return
	}

	// Return the telemetry data
	c.JSON(http.StatusOK, telemetryData)
}

// GetTelemetryHistory handles GET /api/v1/telemetry/:deviceId/history
func (h *SensorHandler) GetTelemetryHistory(c *gin.Context) {
	deviceID := c.Param("deviceId")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device ID is required"})
		return
	}

	// Get query parameters for date range
	from := c.Query("from")
	to := c.Query("to")

	// Fetch telemetry history from the external API
	history, err := h.TelemetryService.GetTelemetryHistory(deviceID, from, to)
	if err != nil {
		if err.Error() == "device with id not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "device with id not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to fetch telemetry history: %v", err),
		})
		return
	}

	// Return the telemetry history
	c.JSON(http.StatusOK, history)
}
