package monitoring

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()

		c.Next()

		status := fmt.Sprintf("%d", c.Writer.Status())
		RequestCount.WithLabelValues(
			"api-service",
			c.Request.Method,
			path,
			status,
		).Inc()

		RequestDuration.WithLabelValues(
			"api-service",
			c.Request.Method,
			path,
		).Observe(time.Since(start).Seconds())
	}
}
