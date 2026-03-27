package main
import (
	"fmt"
	"net/http"
	"time"
	"sync"
    "github.com/gin-gonic/gin"
)
	
func main() { 
	r := gin.New()

	//====	全局中间件	====
	r.Use(loggerMiddleware())
	r.Use(recoverMiddleware())

	//====	路由	====
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	//====	分组中间件	====
	api := r.Group("/api")
	api.Use(authMiddleware())
	{
		api.GET("/user", func(c *gin.Context) {
			userID ,_:= c.Get("userID")
			
			c.JSON(http.StatusOK, gin.H{
				"userID": userID,
			})
		})
	}

	//不需要认证的路由
	r.GET("/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "public route",
		})
	})
	//启动服务
	r.Run(":8080")
}

//日志中间件
func loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("loggerMiddleware")
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		//进入下一个处理函数
		c.Next()

		//后置处理
		latency := time.Since(start)
		status := c.Writer.Status()
		fmt.Printf("%s %s %d %v\n", method, path, status, latency)
	}
}

//恢复中间件
func recoverMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recover interface{}) {
		fmt.Println("recoverMiddleware")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})	
	})
}

//认证中间件
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("authMiddleware")
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}
		if token != "Bearer valid-token" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		//将用户信息存储到请求的上下文中
		c.Set("userID", 1)
		c.Set("username", "admin")

		c.Next()
	}
}

//CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("corsMiddleware")
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		}
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	}
}

//限流中间件
var (
	requestCount = make(map[string]int)
	lastResetTime =time.Now()
	mutex sync.Mutex // 互斥锁
)
var blockedIPs = map[string]bool{
	"127.0.0.1": true,
	"192.168.1.1": true,
}
func rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("rateLimitMiddleware")
		//获取用户IP
		ip := c.ClientIP()
		now := time.Now()

		mutex.Lock()
		defer mutex.Unlock()
		//每分钟重置一次
		if now.Sub(lastResetTime) > time.Minute {
			requestCount = make(map[string]int)
			lastResetTime = now
		}
		//检查请求次数
		if requestCount[ip] >= 10 {
			c.Header("retry-after", "60")// 告知60秒后重试
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Too many requests,please try again later",
			})
			c.Abort()
			return
		}
		//检查IP是否在限制列表中
		if blockedIPs[ip] {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}
		requestCount[ip]++
		c.Next()
	}
}
