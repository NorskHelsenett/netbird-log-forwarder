package routes

import (
	"github.com/NorskHelsenett/netbird-log-forwarder/internal/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(server *gin.Engine) {
	server.POST("/webhook", handlers.RecieveEvent)
}
