package webserver

import (
	"github.com/NorskHelsenett/netbird-log-forwarder/internal/logger"
	"github.com/NorskHelsenett/netbird-log-forwarder/internal/middleware"
	"github.com/NorskHelsenett/netbird-log-forwarder/internal/routes"
	"github.com/gin-gonic/gin"
)

func InitHttpServer() {
	gin.SetMode(gin.ReleaseMode)
	server := gin.New() // or gin.Default()

	token, err := middleware.LoadToken("token.secret")
	if err != nil {
		logger.Log.Fatalf("Failed to load token: %v", err)
	}

	server.Use(gin.Recovery())
	server.Use(middleware.TokenAuthMiddleware(token))
	// server.Use(middleware.ZapLogger())
	// server.Use(middleware.ZapErrorLogger())

	routes.SetupRoutes(server)

	logger.Log.Infoln("NetBird log forwarder starting on port 3000")
	err = server.Run(":3000")

	if err != nil {
		panic(err)
	}
}
