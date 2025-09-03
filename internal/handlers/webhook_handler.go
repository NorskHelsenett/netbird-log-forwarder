package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/NorskHelsenett/netbird-log-forwarder/internal/logger"
	"github.com/NorskHelsenett/netbird-log-forwarder/internal/services"
	"github.com/NorskHelsenett/netbird-log-forwarder/pkg/models/apicontracts"
	"github.com/gin-gonic/gin"
)

type messagePreview struct {
	Message string `json:"message"`
}

func RecieveEvent(ginContext *gin.Context) {
	requestBody, err := io.ReadAll(ginContext.Request.Body)
	if err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": "could not read request body"})
		return
	}
	ginContext.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))

	var preview messagePreview
	_ = json.Unmarshal(requestBody, &preview)

	switch {
	case looksLikeTrafficEvent(preview.Message):
		logger.Log.Debugln("Processing traffic event")
		var event apicontracts.TrafficEvent
		eventBody := json.NewDecoder(bytes.NewReader(requestBody))
		eventBody.DisallowUnknownFields()

		if err := eventBody.Decode(&event); err != nil {
			ginContext.JSON(http.StatusBadRequest, gin.H{"message": "invalid traffic payload", "error": err.Error()})
			return
		}

		_, err := services.ProcessTrafficEvent(event)
		if err != nil {
			ginContext.Error(err)
			ginContext.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		// logger.SplunkTraffic.Infow("traffic_event", "event", evt)
		ginContext.JSON(http.StatusAccepted, gin.H{"status": "ok", "handled_as": "traffic"})
		return
	case looksLikeAuditEvent(preview.Message):
		logger.Log.Debugln("Processing audit event")

		// Log the request body in pretty JSON format
		var prettyBody bytes.Buffer
		if err := json.Indent(&prettyBody, requestBody, "", "  "); err != nil {
			logger.Log.Errorf("Failed to pretty-print request body: %v", err)
		} else {
			logger.Log.Infof("Received request body:\n%s", prettyBody.String())
		}

		var event apicontracts.AuditEvent
		eventBody := json.NewDecoder(bytes.NewReader(requestBody))
		// eventBody.DisallowUnknownFields()

		if err := eventBody.Decode(&event); err != nil {
			ginContext.JSON(http.StatusBadRequest, gin.H{"message": "invalid audit payload", "error": err.Error()})
			return
		}

		_, err := services.ProcessAuditEvent(event)
		if err != nil {
			ginContext.Error(err)
			ginContext.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		// logger.SplunkTraffic.Infow("traffic_event", "event", evt)
		ginContext.JSON(http.StatusAccepted, gin.H{"status": "ok", "handled_as": "audit"})
		return
	default:
		logger.Log.Debugln("Unknown event type")
		var event any
		eventBody := json.NewDecoder(bytes.NewReader(requestBody))
		// eventBody.DisallowUnknownFields()
		if err := eventBody.Decode(&event); err != nil {
			ginContext.JSON(http.StatusBadRequest, gin.H{"message": "invalid audit payload", "error": err.Error()})
			return
		}
		prettyJSON, err := json.MarshalIndent(event, "", "  ")
		if err != nil {
			logger.Log.Errorf("Failed to marshal event to pretty JSON: %v", err)
		} else {
			logger.Log.Infof("Unknown event type:\n%s", prettyJSON)

		}
	}

	ginContext.JSON(http.StatusOK, gin.H{"message": "Event processed successfully"})

}

func looksLikeTrafficEvent(msg string) bool {
	switch strings.ToUpper(strings.TrimSpace(msg)) {
	case "TYPE_START", "TYPE_END", "TYPE_DROP":
		return true
	default:
		return false
	}
}

func looksLikeAuditEvent(msg string) bool {
	switch strings.TrimSpace(msg) {
	case "user blocked", "user unblocked", "peer approved":
		return true
	default:
		return false
	}
}
