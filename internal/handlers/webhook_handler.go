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
			if err.Error() == "not_splunk_worthy" {
				ginContext.JSON(http.StatusAccepted, gin.H{"status": "ok"})
				return
			}
			ginContext.Error(err)
			ginContext.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		ginContext.JSON(http.StatusAccepted, gin.H{"status": "ok", "handled_as": "traffic"})
		return
	default:
		logger.Log.Debugln("Processing audit event")
		var event apicontracts.AuditEventEnvelope

		if err := json.Unmarshal(requestBody, &event); err != nil {
			logger.SplunkAudit.Errorw("failed to decode", "error", err)
			return
		}
		_, err = services.ProcessAuditEvent(event)
		if err != nil {

			ginContext.Error(err)
			ginContext.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		ginContext.JSON(http.StatusAccepted, gin.H{"status": "ok", "handled_as": "audit"})
		return

	}
}

func looksLikeTrafficEvent(msg string) bool {
	return strings.HasPrefix(strings.ToUpper(strings.TrimSpace(msg)), "TYPE_")
}
