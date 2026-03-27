package main
import (
	"fmt"
	"net/http"
    "github.com/gin-gonic/gin"
)

func main() { 
	r := gin.Default()

	//基础路由
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World",
		})
	})

	//单个参数
	r.GET("/user/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello " + name,
		})
	})

	//多个参数
	r.GET("/user/:name/:age", func(c *gin.Context) {
		name := c.Param("name")
		age := c.Param("age")
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello " + name + " " + age,
		})
	})
	//通配符
	r.GET("/files/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello " + filepath,
		})
	})

	//查询参数
	r.GET("/search", func(c *gin.Context) {
		name := c.DefaultQuery("name","爱丽丝")
		age := c.DefaultQuery("age","22")
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello    姓名:" + name + "  年龄：" + age,
		})
	})

	//HTML表单
	r.POST("/form", func(c *gin.Context) {
		title := c.PostForm("title")
		content := c.PostForm("content")
		author := c.DefaultPostForm("author","alice")
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello    标题:" + title + "  正文：" + content + "  作者：" + author,
		})
	})

	//json参数
	r.POST("/jsonP", func(c *gin.Context) {
		type jParam struct {
			Title   string `json:"title"`
			Content string `json:"content"`
			Author  string `json:"author"`
		}
		var req jParam
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello    标题:" + req.Title + "  正文：" + req.Content + "  作者：" + req.Author,
		})
	})

	//基础分组
	v1 := r.Group("/api/v1")
	{
		v1.GET("/users", getUsers)
		v1.GET("/users/:id", getUser)
		v1.POST("/users", createUser)
	}

	//嵌套分组
	api := r.Group("/api")
	{
		v2 := api.Group("/v2")
		{
			v2.GET("/users", getUsers)
		}
		v3 := api.Group("/v3")
		{
			v3.GET("/users", getUsers)
			v3.Any("/like", likePerson)
		}
	}


	//启动服务
	r.Run(":8080")
}
func likePerson(c *gin.Context){
	c.JSON(http.StatusOK, gin.H{
		"message": "喜欢这个人",
	})
}
func getUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "获取到所有用户",
	})
}
func getUser(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "获取到用户" + id,
	})
}
func createUser(c *gin.Context) {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	msg := fmt.Sprintf("创建用户 %s 成功,年龄 %d 岁", user.Name,user.Age)
	c.JSON(http.StatusOK, gin.H{
		"message": msg,
	})
}