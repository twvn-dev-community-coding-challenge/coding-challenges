package routes

import (
	"sms-service/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine, smsHandler *handlers.SMSMessageHandler) {
	router.GET("/health", handlers.HealthCheck)

	v1 := router.Group("/api/v1")
	{
		sms := v1.Group("/sms")
		sms.POST("/send", smsHandler.SendSMS)
		sms.POST("/webhooks/provider-callback", smsHandler.HandleProviderCallback)
		sms.GET("/:id", smsHandler.GetByID)
	}
}
