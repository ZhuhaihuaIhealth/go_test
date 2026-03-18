package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"BlogSystem/testutil"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Claims JWT 自定义声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func main() {
	testutil.LoadEnv()

	db, err := testutil.OpenDB("blog.db")
	if err != nil {
		log.Fatal("数据库连接失败: ", err)
	}

	if err := db.AutoMigrate(&User{}, &Post{}, &Comment{}); err != nil {
		log.Fatal("数据库迁移失败: ", err)
	}

	r := gin.Default()

	templateDir := getTemplateDir()
	r.LoadHTMLGlob(filepath.Join(templateDir, "*.html"))

	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", gin.H{})
	})

	r.POST("/register", rateLimitMiddleware(), func(c *gin.Context) {
		username := c.PostForm("username")
		email := c.PostForm("email")
		password := c.PostForm("password")

		_, err := RegisterUser(db, username, password, email)
		if err != nil {
			c.HTML(http.StatusBadRequest, "register.html", gin.H{
				"Message": err.Error(),
				"Success": false,
			})
			return
		}

		c.Redirect(http.StatusFound, "/login?msg=注册成功，请登录")
	})

	r.GET("/login", func(c *gin.Context) {
		data := gin.H{}
		if msg := c.Query("msg"); msg != "" {
			data["Message"] = msg
			data["Success"] = true
		}
		c.HTML(http.StatusOK, "login.html", data)
	})

	r.POST("/login", rateLimitMiddleware(), func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		tokenStr, err := login(db, username, password)
		if err != nil {
			c.HTML(http.StatusUnauthorized, "login.html", gin.H{
				"Message": "登录失败，" + err.Error(),
				"Success": false,
			})
			return
		}

		c.SetCookie("token", tokenStr, 86400, "/", "", false, false)
		c.Header("Authorization", "Bearer "+tokenStr)
		c.Redirect(http.StatusFound, "/api/profile")
	})

	r.GET("/logout", func(c *gin.Context) {
		c.SetCookie("token", "", -1, "/", "", false, true)
		c.Redirect(http.StatusFound, "/login?msg=已退出登录")
	})

	// 需要认证的路由组
	auth := r.Group("/api")
	auth.Use(rateLimitMiddleware(), AuthMiddleware())
	{
		auth.GET("/profile", func(c *gin.Context) {
			username, _ := c.Get("username")
			userID, _ := c.Get("userID")
			posts, _ := GetUserPosts(db, userID.(uint))
			c.HTML(http.StatusOK, "profile.html", gin.H{
				"Username": username,
				"UserID":   userID,
				"Posts":    posts,
			})
		})

		auth.GET("/posts", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			posts, err := GetUserPosts(db, userID.(uint))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"posts": posts})
		})

		auth.GET("/posts/all", func(c *gin.Context) {
			posts, err := GetAllPosts(db)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"posts": posts})
		})

		auth.POST("/posts", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			var req struct {
				Title   string `json:"title"`
				Content string `json:"content"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
				return
			}
			post, err := CreateUserPost(db, userID.(uint), req.Title, req.Content)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"post": post})
		})

		auth.DELETE("/posts/:id", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			postID, err := strconv.ParseUint(c.Param("id"), 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文章ID"})
				return
			}
			if err := DeleteUserPost(db, userID.(uint), uint(postID)); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
		})

		auth.POST("/posts/:id/comments", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			postID, err := strconv.ParseUint(c.Param("id"), 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文章ID"})
				return
			}
			var req struct {
				Content string `json:"content"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
				return
			}
			comment, err := CreateComment(db, userID.(uint), uint(postID), req.Content)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"comment": comment})
		})

		auth.GET("/posts/:id", func(c *gin.Context) {
			postID, err := strconv.ParseUint(c.Param("id"), 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文章ID"})
				return
			}
			post, comments, err := GetPostDetail(db, uint(postID))
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"post": post, "comments": comments})
		})
	}

	log.Println("服务启动: http://localhost:8080/register")
	r.Run(":8080")
}

// login 验证用户名密码，成功返回 JWT token
func login(db *gorm.DB, username, password string) (string, error) {
	user, err := GetUser(db, username)
	if err != nil {
		return "", errors.New("用户名或密码错误")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("用户名或密码错误")
	}
	return generateToken(user.ID, username)
}

func getJWTKey() []byte {
	key := os.Getenv("JWT_KEY")
	if key == "" {
		log.Fatal("JWT_KEY 环境变量未设置")
	}
	return []byte(key)
}

// generateToken 生成 JWT token
func generateToken(userID uint, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTKey())
}

// parseToken 解析并验证 JWT token
func parseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return getJWTKey(), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("无效的 token")
}

// AuthMiddleware JWT 认证中间件，优先从 Authorization Header 读取，fallback 到 Cookie
func AuthMiddleware() gin.HandlerFunc {
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

		claims, err := parseToken(tokenStr)
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
		c.Next()
	}
}

func getTemplateDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("获取工作目录失败: ", err)
	}
	return filepath.Join(dir, "templates")
}

func formatUint(n uint) string {
	if n == 0 {
		return "0"
	}
	digits := make([]byte, 0, 10)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return string(digits)
}
// 限流中间件
var (
	requestCount  = make(map[string]int)
	failCount     = make(map[string]int)
	lastResetTime = time.Now()
	blockedIPs    = make(map[string]time.Time)
	mutex         sync.Mutex
)

func rateLimitMiddleware() gin.HandlerFunc {
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
