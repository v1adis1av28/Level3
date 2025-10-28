package middleware

import (
	"time"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

func LoggingMiddleware() ginext.HandlerFunc {
	return func(c *ginext.Context) {
		CallTime := time.Now()
		c.Next()
		method := c.Request.Method
		url := c.Request.URL.Path
		zlog.Logger.Info().Msgf("Method : %s, url : %s, Time: %v", method, url, CallTime)
		c.Next()
	}
}
