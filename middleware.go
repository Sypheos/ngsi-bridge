// Copyright © 2018 The Things Industries B.V.

package bridges

import (
	"fmt"
	"net/http"
	"time"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/gin-gonic/gin"
)

// Logger middleware for http.
func Logger(ctx log.Interface) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		end := time.Now()

		version := c.GetHeader("X-Version")
		if version == "" {
			if v := c.Query("version"); v != "" {
				version = v
			} else {
				version = "unknown"
			}
		}
		fields := log.Fields{
			"Method":        c.Request.Method,
			"Path":          c.Request.URL.Path,
			"Query":         c.Request.URL.Query(),
			"Host":          c.Request.Host,
			"Id":            c.GetHeader("X-Request-ID"),
			"RemoteAddress": c.ClientIP(),
			"Duration":      end.Sub(start),
			"Status":        c.Writer.Status(),
			"Version":       version,
		}
		if err := c.Errors.ByType(gin.ErrorTypePrivate).Last(); err != nil {
			ctx = ctx.WithError(err)
		}
		if err := c.Errors.ByType(gin.ErrorTypePublic).Last(); err != nil {
			fmt.Println(err)
		}
		ctx.WithFields(fields).Info("Inbound request")
	}
}

func checkHeaders(c *gin.Context) {
	if c.GetHeader("Content-Type") != "application/json" {
		c.Header("Allowed-Content-Type", "application/json")
		c.Status(http.StatusBadRequest)
		c.Abort()
	}
}

func jsonError(c *gin.Context) {
	c.Next()
	lastError := c.Errors.ByType(gin.ErrorTypePublic).Last()
	if lastError == nil {
		return
	}
	c.JSON(-1, lastError)
}
