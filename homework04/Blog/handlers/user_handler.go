package handlers

import (
	"net/http"
	"strconv"

	"BlogSystem/middleware"
	"BlogSystem/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", gin.H{})
	})

	r.POST("/register", middleware.RateLimit(), func(c *gin.Context) {
		username := c.PostForm("username")
		email := c.PostForm("email")
		password := c.PostForm("password")

		_, err := services.RegisterUser(db, username, password, email)
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

	r.POST("/login", middleware.RateLimit(), func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		tokenStr, err := services.Login(db, username, password)
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

	auth := r.Group("/api")
	auth.Use(middleware.RateLimit(), middleware.Auth())
	{
		auth.GET("/profile", func(c *gin.Context) {
			username, _ := c.Get("username")
			userID, _ := c.Get("userID")
			token, _ := c.Get("token")
			posts, _ := services.GetUserPosts(db, userID.(uint))
			c.HTML(http.StatusOK, "profile.html", gin.H{
				"Username": username,
				"UserID":   userID,
				"Posts":    posts,
				"Token":    token,
			})
		})

		auth.GET("/posts", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			posts, err := services.GetUserPosts(db, userID.(uint))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"posts": posts})
		})

		auth.GET("/posts/all", func(c *gin.Context) {
			posts, err := services.GetAllPosts(db)
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
			post, err := services.CreateUserPost(db, userID.(uint), req.Title, req.Content)
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
			if err := services.DeleteUserPost(db, userID.(uint), uint(postID)); err != nil {
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
			comment, err := services.CreateComment(db, userID.(uint), uint(postID), req.Content)
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
			post, comments, err := services.GetPostDetail(db, uint(postID))
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"post": post, "comments": comments})
		})
	}
}
