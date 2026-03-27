# 作业要求概述
本次作业要求你使用 Go 语言结合 Gin 框架和 GORM 库开发一个个人博客系统的后端，实现博客文章的基本管理功能，包括文章的创建、读取、更新和删除（CRUD）操作，同时支持用户认证和简单的评论功能。

## 1. 项目初始化
- 创建一个新的 Go 项目，使用 go mod init 初始化项目依赖管理。
- 安装必要的库，如 Gin 框架、GORM 以及数据库驱动（如 MySQL 或 SQLite）。


## 2. 数据库设计与模型定义
- 设计数据库表结构，至少包含以下几个表：
  - users 表：存储用户信息，包括 id 、 username 、 password 、 email 等字段。
  - posts 表：存储博客文章信息，包括 id 、 title 、 content 、 user_id （关联 users 表的 id ）、 created_at 、 updated_at 等字段。
  - comments 表：存储文章评论信息，包括 id 、 content 、 user_id （关联 users 表的 id ）、 post_id （关联 posts 表的 id ）、 created_at 等字段。
- 使用 GORM 定义对应的 Go 模型结构体。

## 3. 用户认证与授权
- 实现用户注册和登录功能，用户注册时需要对密码进行加密存储，登录时验证用户输入的用户名和密码。
- 使用 JWT（JSON Web Token）实现用户认证和授权，用户登录成功后返回一个 JWT，后续的需要认证的接口需要验证该 JWT 的有效性。

## 4. 文章管理功能
- 实现文章的创建功能，只有已认证的用户才能创建文章，创建文章时需要提供文章的标题和内容。
- 实现文章的读取功能，支持获取所有文章列表和单个文章的详细信息。
- 实现文章的更新功能，只有文章的作者才能更新自己的文章。
- 实现文章的删除功能，只有文章的作者才能删除自己的文章。

## 5. 评论功能
- 实现评论的创建功能，已认证的用户可以对文章发表评论。
- 实现评论的读取功能，支持获取某篇文章的所有评论列表。

## 6. 错误处理与日志记录
- 对可能出现的错误进行统一处理，如数据库连接错误、用户认证失败、文章或评论不存在等，返回合适的 HTTP 状态码和错误信息。
- 使用日志库记录系统的运行信息和错误信息，方便后续的调试和维护。

## 代码示例参考

数据库连接与模型定义
```go
package main

import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

type User struct {
    gorm.Model
    Username string `gorm:"unique;not null"`
    Password string `gorm:"not null"`
    Email    string `gorm:"unique;not null"`
}

type Post struct {
    gorm.Model
    Title   string `gorm:"not null"`
    Content string `gorm:"not null"`
    UserID  uint
    User    User
}

type Comment struct {
    gorm.Model
    Content string `gorm:"not null"`
    UserID  uint
    User    User
    PostID  uint
    Post    Post
}

func main() {
    db, err := gorm.Open(sqlite.Open("blog.db"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }

    // 自动迁移模型
    db.AutoMigrate(&User{}, &Post{}, &Comment{})
}
 ```

用户注册与登录示例
```go
package main

import (
    "github.com/dgrijalva/jwt-go"
    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "net/http"
    "time"
)

func Register(c *gin.Context) {
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    // 加密密码
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }
    user.Password = string(hashedPassword)

    if err := db.Create(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func Login(c *gin.Context) {
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var storedUser User
    if err := db.Where("username = ?", user.Username).First(&storedUser).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
        return
    }

    // 验证密码
    if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password)); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
        return
    }

    // 生成 JWT
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "id":       storedUser.ID,
        "username": storedUser.Username,
        "exp":      time.Now().Add(time.Hour * 24).Unix(),
    })

    tokenString, err := token.SignedString([]byte("your_secret_key"))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }
    // 剩下的逻辑...
}
 ```

## 提交要求
- 提交完整的项目代码，包括必要的配置文件和依赖管理文件。
- 提供项目的 README 文件，说明项目的运行环境、依赖安装步骤和启动方式。
- 可以使用 Postman 或其他工具对接口进行测试，并提供测试用例和测试结果。
