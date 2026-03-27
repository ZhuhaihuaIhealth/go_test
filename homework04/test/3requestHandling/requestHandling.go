package main
import (
	// "fmt"
	"net/http"
    "github.com/gin-gonic/gin"
)

//请求结构体
type CreateUserRequest struct {
    Name string `json:"name"`
    Email string `json:"email" binding:"required,email"`
}
type UpdateUserRequest struct {
    Name string `json:"name"`
	Email string `json:"email" binding:"required,email"`
}
type ListProductRequest struct {
	Page int `form:"page" binding:"gte=1"`
	PageSize int `form:"size" binding:"gte=1,lte=100"`
	Keyword string `form:"keyword"`
}
type LoginRequest struct {
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
type GetUserRequest struct {
	ID uint `uri:"id" binding:"required"`
}
//响应结构体
type UserResponse struct {
	Code int `json:"code"`
	Message string `json:"message"`
	Data interface{} `json:"data",omitempty`
}

func success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, UserResponse{
		Code: 200,
		Message: "success",
		Data: data,
	})
}
func fail(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, UserResponse{
		Code: code,
		Message: message,
	})
}

func main() {
	r := gin.Default()

	//JSON绑定
	r.POST("/api/users", func(c *gin.Context) { 
		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			fail(c, http.StatusBadRequest, err.Error())
			return
		}
		success(c, gin.H{
			"name": req.Name,
			"email": req.Email,
		})
	})

	//查询参数绑定
	r.GET("/api/products", func(c *gin.Context) { 
		var req ListProductRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			fail(c, http.StatusBadRequest, err.Error())
			return
		}
		success(c, gin.H{
			"page": req.Page,
			"page_size": req.PageSize,
			"keyword": req.Keyword,
			"products": []gin.H{
				{"id":1,"name": "product1"},
				{"id":2,"name": "product2"},
			},
		})
	})

	//路径参数绑定
	r.GET("/api/users/:id", func(c *gin.Context) { 
		var req GetUserRequest
		if err := c.ShouldBindUri(&req); err != nil {
			fail(c, http.StatusBadRequest, err.Error())
			return
		}
		success(c, gin.H{
			"id": req.ID,
		})
	})

	//表单绑定
	r.POST("/api/login", func(c *gin.Context) { 
		var req LoginRequest
		if err := c.ShouldBind(&req); err != nil {
			fail(c, http.StatusBadRequest, err.Error())
			return
		}
		//模拟登录验证
		if req.Email != "admin@example.com" || req.Password != "123456" {
			fail(c, http.StatusUnauthorized, "用户名或密码错误")
			return
		}
		success(c, gin.H{
			"email": req.Email,
			"password": req.Password,
		})
	})

	//原始请求体
	r.POST("/api/raw", func(c *gin.Context) { 
		data, err := c.GetRawData()
		if err != nil {
			fail(c, http.StatusBadRequest, err.Error())
			return
		}
		success(c, gin.H{
			"body": string(data),
		})
		
	})

	//多种绑定方式
	r.POST("/api/users/mixed/:id", func(c *gin.Context) { 
		id:= c.Param("id")
		page := c.DefaultQuery("page","1")

		var body CreateUserRequest
		if err := c.ShouldBind(&body); err != nil {
			fail(c, http.StatusBadRequest, err.Error())
			return
		}
		success(c, gin.H{
			"id": id,
			"page": page,
			"name": body.Name,
			"email": body.Email,
		})
	})

	//响应类型
	//JSON响应
	r.GET("/api/json", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "json 响应",
		})
	})
	//xml响应
	r.GET("/api/xml", func(c *gin.Context) {
		c.XML(http.StatusOK, gin.H{
			"message": "xml 响应",
		})
	})
	//字符串响应
	r.GET("/api/string", func(c *gin.Context) {
		c.String(http.StatusOK, "字符串响应")
	})
	//重定向
	r.GET("/api/redirect", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/api/json")
	})
	//文件响应
	r.GET("/api/file", func(c *gin.Context) {
		c.File("./file.txt")
	})
	//数据流响应
	r.GET("/api/stream", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/octet-stream", []byte("数据流响应"))
	})

	r.Run(":8080")
}