package middleware

import (
	"time"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

func LoggingMiddleware() ginext.HandlerFunc {
	return func(c *ginext.Context) {
		start := time.Now()

		c.Next()
		method := c.Request.Method
		url := c.Request.URL.Path

		zlog.Logger.Debug().Fields(map[string]interface{}{
			"method":    method,
			"url":       url,
			"status":    c.Writer.Status(),
			"call_time": time.Since(start).Milliseconds(),
		}).Msg("Request")
	}
}
