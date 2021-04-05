package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var TokenAuthKey = "TokenAuth"

func (s *CCServer) tokenAuth(expectedToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeaderFields := strings.Fields(c.GetHeader("Authorization"))
		if len(authHeaderFields) != 2 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		token := authHeaderFields[1]
		if token == expectedToken {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}
