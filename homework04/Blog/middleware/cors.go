package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func Gzip() gin.HandlerFunc {
	return gzip.Gzip(gzip.DefaultCompression)
}

func StaticCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/static/") {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		}
		c.Next()
	}
}

var (
	requestCount  = make(map[string]int)
	failCount     = make(map[string]int)
	lastResetTime = time.Now()
	blockedIPs    = make(map[string]time.Time)
	mutex         sync.Mutex
)

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("rateLimitMiddleware")
		ip := c.ClientIP()
		now := time.Now()

		mutex.Lock()
		defer mutex.Unlock()

		if blockedAt, ok := blockedIPs[ip]; ok {
			if now.Sub(blockedAt) < 30*time.Minute {
				remaining := 30*time.Minute - now.Sub(blockedAt)
				c.Header("Retry-After", strconv.Itoa(int(remaining.Seconds())))
				c.JSON(http.StatusForbidden, gin.H{
					"error": fmt.Sprintf("IP 已被封禁，%.0f 分钟后解封", remaining.Minutes()),
				})
				c.Abort()
				return
			}
			delete(blockedIPs, ip)
			delete(failCount, ip)
		}

		if now.Sub(lastResetTime) > time.Minute {
			requestCount = make(map[string]int)
			failCount = make(map[string]int)
			lastResetTime = now
		}

		if requestCount[ip] >= 10 {
			c.Header("Retry-After", "60")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}

		requestCount[ip]++
		c.Next()

		if c.Writer.Status() >= 400 {
			failCount[ip]++
			if failCount[ip] >= 5 {
				blockedIPs[ip] = now
				fmt.Printf("IP %s 因1分钟内失败 %d 次被封禁30分钟\n", ip, failCount[ip])
			}
		}
	}
}
