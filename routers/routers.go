package routers

import (
	"github.com/gin-gonic/gin"
	"os"
	"time"

	"email-auth/controllers"
	"email-auth/middleware"
)

func RouterEmailHelper() *gin.Engine {
	r := gin.Default()
	gin.SetMode(os.Getenv("GIN_MOD"))

	r.Use(middleware.RateLimitMiddleware(time.Second, 5, 5))

	t := r.Group("/email")
	{
		t.POST("/send", controllers.SendCode)
		t.POST("/verify", controllers.VerifyEmailCode)
		t.POST("/resend", controllers.ReSendCode)
	}

	s := r.Group("/auth")
	{
		s.GET("/code", controllers.GetAuthCode)
		s.GET("/resource", controllers.GetResource)
		s.PUT("/refresh", controllers.RefreshTokenByRF)
		s.GET("/access", controllers.GetTokenByCode)
	}

	return r
}
