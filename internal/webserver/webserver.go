package webserver

import (
	"github.com/NorskHelsenett/netbird-log-forwarder/internal/logger"
	"github.com/NorskHelsenett/netbird-log-forwarder/internal/middleware"
	"github.com/NorskHelsenett/netbird-log-forwarder/internal/routes"
	"github.com/gin-gonic/gin"
)

func InitHttpServer(token string) {
	gin.SetMode(gin.ReleaseMode)
	server := gin.New() // or gin.Default()

	// token := viper.GetString("api.auth_token")
	server.Use(gin.Recovery())
	server.Use(middleware.TokenAuthMiddleware(token))
	routes.SetupRoutes(server)

	logger.Log.Infoln("NetBird log forwarder starting on port 3000")
	err := server.Run(":3000")

	if err != nil {
		panic(err)
	}
}
