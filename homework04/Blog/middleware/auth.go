package middleware

import (
	"net/http"
	"strings"

	"BlogSystem/utils"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenStr string

		if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			tokenStr = strings.TrimPrefix(auth, "Bearer ")
		}

		if tokenStr == "" {
			tokenStr, _ = c.Cookie("token")
		}

		if tokenStr == "" {
			if c.GetHeader("Accept") == "application/json" || strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录，请先登录"})
			} else {
				c.Redirect(http.StatusFound, "/login?msg=请先登录")
			}
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(tokenStr)
		if err != nil {
			c.SetCookie("token", "", -1, "/", "", false, true)
			if c.GetHeader("Accept") == "application/json" || strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "登录已过期，请重新登录"})
			} else {
				c.Redirect(http.StatusFound, "/login?msg=登录已过期，请重新登录")
			}
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("token", tokenStr)
		c.Next()
	}
}
