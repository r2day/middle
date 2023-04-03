package middle

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// RequestRespMiddleware 系统调用日志输出
func RequestRespMiddleware(enableLog bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		if(enableLog) {
			log.WithFields(log.Fields{
				"method": c.Request.Method,
				"path": c.FullPath(),
				"status": c.Writer.Status(),
				"latency_ns": time.Since(start).Nanoseconds()
			}).Info("request_detail")
		}
	}
}