package main
import(
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
	"os"
	"fmt"
	"sync"
	"runtime"
	"strings"
	"path/filepath"
	"github.com/joho/godotenv"
)

// var jwtKey = []byte(os.Getenv("JWT_KEY"))
type Claims struct {
	UserID uint `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}
func main() { 
	// LoadEnv()

	r := gin.Default()

	//公开路由
	r.POST("/api/login", login)

	r.GET("/api/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "public endpoint",
		})
	})

	//需要认证的路由
	api := r.Group("/api")
	api.Use(authMiddleware())
	{
		api.GET("/protected", func(c *gin.Context) {
			userID ,_:= c.Get("userID")
			username ,_:= c.Get("username")
			
			c.JSON(http.StatusOK, gin.H{
				"userID": userID,
				"username": username,
				"message": "protected 返回的内容",
			})
		})
		api.GET("/profile", getProfile)
	}

	//启动服务
	r.Run(":8080")
}
var (
	envLoaded bool
	envOnce   sync.Once
)
func LoadEnv() {
	envOnce.Do(func() {

		//获取当前文件
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			panic("无法获取当前文件")
		}
		//获取当前文件所在目录
		testUtilDir := filepath.Dir(filename)
		// homeworkDir := filepath.Dir(testUtilDir)
		envfile := filepath.Join(testUtilDir, ".env")
		if err := godotenv.Load(envfile); err != nil {
			// panic("无法加载环境变量文件")
			return
		}
		envLoaded = true
	})
}

func generateToken(userID uint, username string)(string, error){
	LoadEnv()
	var jwtKey = []byte(os.Getenv("JWT_KEY"))
	claims := Claims{
		UserID: userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
func parseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _,ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(os.Getenv("JWT_KEY")), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	} 
	return nil, errors.New("invalid token")
}
func login(c *gin.Context){
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//验证用户名密码（示例，实际应从数据库查询）
	if req.Username == "admin" || req.Password == "123456" {
		token,err := generateToken(1, req.Username)
		if err != nil {
			fmt.Println("生成token失败")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{
				"id": 1,
				"username": req.Username,
			},
		})
		return
	}
	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
}
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

		//提取token
		parts := strings.Split(token, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token format",
			})
			c.Abort()
			return
		}
		tokenString := parts[1]

		//验证token
		claims, err := parseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}
		
		//将用户信息存储到contex
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
func getProfile(c *gin.Context){
	userID ,_:= c.Get("userID")
	username ,_:= c.Get("username")
	c.JSON(http.StatusOK, gin.H{
		"userID": userID,
		"username": username,
		"message": "profile endpoint",
	})
}