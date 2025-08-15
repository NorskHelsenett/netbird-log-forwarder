package handlers

import (
	// "net/http"

	"fmt"
	"net/http"

	// "github.com/NorskHelsenett/netbird-log-forwarder/internal/services"
	// "github.com/NorskHelsenett/netbird-log-forwarder/pkg/models/apicontracts"
	"github.com/NorskHelsenett/netbird-log-forwarder/internal/logger"
	"github.com/gin-gonic/gin"
)

func RecieveEvent(ginContext *gin.Context) {
	var request any //apicontracts.TrafficEvent
	err := ginContext.ShouldBindJSON(&request)

	if err != nil {
		ginContext.Error(err)
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse incomming request"})
		fmt.Println(err)
		return
	}

	// logger.Log.Infow("Received request", "request", request)
	logger.Log.Infow("Received request", "request", request)
	// jsonBytes, err := json.MarshalIndent(request, "", "  ")
	// if err != nil {
	// 	fmt.Println("Failed to marshal request:", err)
	// } else {
	// 	fmt.Println("Request received:\n", string(jsonBytes))
	// }

	// var response any
	httpStatus := http.StatusOK

	// response, err = services.ProcessTrafficEvent(request)
	// if err != nil {
	// 	ginContext.Error(err)
	// 	ginContext.JSON(http.StatusInternalServerError, gin.H{"message": "ERROR: " + err.Error()})
	// 	return
	// }

	ginContext.JSON(httpStatus, gin.H{"message": "Event processed successfully"})

}
