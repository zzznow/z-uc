package common

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("traceId", xid.New().String())
		c.Next()
	}
}
