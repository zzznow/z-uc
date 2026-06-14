package common

import (
	"github.com/gin-gonic/gin"
)

const ServiceNameCtxKey = "svcName"

func SetServiceName(c *gin.Context, name string) {
	c.Set(ServiceNameCtxKey, name)
}

func svcName(c *gin.Context) string {
	if v, ok := c.Get(ServiceNameCtxKey); ok {
		return v.(string)
	}
	return "unknown"
}

func JSON(c *gin.Context, code int, obj interface{}) {
	c.AbortWithStatusJSON(code, gin.H{"data": obj})
}

func Error(c *gin.Context, code int, err *AppError) {
	if err.TraceID == "" {
		if v, ok := c.Get("traceId"); ok {
			err.TraceID = v.(string)
		}
	}
	if err.Chain == "" {
		err.Chain = svcName(c)
	} else {
		err.Chain = svcName(c) + " <- " + err.Chain
	}
	c.AbortWithStatusJSON(code, gin.H{"error": err})
}

func ErrorMsg(c *gin.Context, code int, msg string) {
	Error(c, code, NewError(msg))
}
