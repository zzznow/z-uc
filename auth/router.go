package auth

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.String(200, "ok")
	})
	r.POST("/login", FormLogin)
	r.POST("/login/refresh", RefreshToken)
	r.POST("/login/sms", SmsLogin)
	r.POST("/register", SignUp)
	r.POST("/auth/state", NewState)
	r.POST("/auth/google/token", GoogleToken)
	r.POST("/auth/wx/token", WxToken)
	r.POST("/auth/wx-miniapp/token", WxMiniToken)
	r.GET("/auth/token/verify", VerifyTokenHandler)
	r.GET("/auth/info", GetTokenInfo)
}
