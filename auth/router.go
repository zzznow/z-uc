package auth

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	r.GET("/healthz", func(c *gin.Context) {
		c.String(200, "ok")
	})

	// ── 公开路由（无需 Token）────────────────────
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

	// ── 用户自有路由（需 z-uc Bearer Token）──────
	auth := r.Group("/auth", AuthMiddleware())
	{
		auth.GET("/profile", GetProfile)
		auth.PUT("/profile", UpdateProfile)
		auth.POST("/password/change", ChangePassword)
		auth.POST("/bind/phone", BindPhone)
		auth.POST("/bind/username", BindUsername)
		auth.POST("/cancel", CancelAccount)
	}

	// ── 内部 API（服务间调用，内网可达）────────────
	internal := r.Group("/internal/auth")
	{
		internal.POST("/bind-phone", InternalBindPhone)
		internal.POST("/bind-username", InternalBindUsername)
		internal.POST("/change-password", InternalChangePassword)
		internal.GET("/profile/:sn", InternalGetProfile)
		internal.POST("/verify-token", InternalVerifyToken)

		// 已有的内部接口
		internal.GET("/user/sn/:sn", GetUserBySnInternal)
		internal.GET("/user/unionId/:unionId", GetUserByUnionIdInternal)
		internal.GET("/user/id", GetUserIdInternal)
	}
}
