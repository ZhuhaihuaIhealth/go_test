package main

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ==================== 题目1：模型定义 ====================

type User struct {
	gorm.Model
	Name      string `gorm:"size:100;not null"`
	Email     string `gorm:"size:200;uniqueIndex;not null"`
	PostCount int    `gorm:"default:0"`
	Posts     []Post `gorm:"foreignKey:UserID"`
}

type Post struct {
	gorm.Model
	Title         string    `gorm:"size:200;not null"`
	Content       string    `gorm:"type:text"`
	UserID        uint      `gorm:"not null;index"`
	CommentStatus string    `gorm:"size:50;default:无评论"`
	Comments      []Comment `gorm:"foreignKey:PostID"`
}

type Comment struct {
	gorm.Model
	Content string `gorm:"type:text;not null"`
	PostID  uint   `gorm:"not null;index"`
}

// ==================== 题目3：钩子函数 ====================

// AfterCreate 文章创建后自动更新用户的文章数量统计
func (p *Post) AfterCreate(tx *gorm.DB) error {
	return tx.Model(&User{}).Where("id = ?", p.UserID).
		UpdateColumn("post_count", gorm.Expr("post_count + 1")).Error
}

// AfterDelete 评论删除后检查文章评论数量，为0则标记"无评论"
func (c *Comment) AfterDelete(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&Comment{}).Where("post_id = ?", c.PostID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return tx.Model(&Post{}).Where("id = ?", c.PostID).
			UpdateColumn("comment_status", "无评论").Error
	}
	return nil
}

func main() {
	db, err := gorm.Open(sqlite.Open("blog.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 题目1：创建数据库表
	if err := db.AutoMigrate(&User{}, &Post{}, &Comment{}); err != nil {
		log.Fatal("建表失败:", err)
	}
	fmt.Println("=== 数据库表创建成功 ===")

	// 插入测试数据
	seedData(db)

	// ==================== 题目2：关联查询 ====================

	// 查询某个用户发布的所有文章及其对应的评论
	queryUserPostsWithComments(db)

	// 查询评论数量最多的文章
	queryMostCommentedPost(db)

	// 演示题目3的钩子效果
	demonstrateHooks(db)
}

func seedData(db *gorm.DB) {
	var count int64
	db.Model(&User{}).Count(&count)
	if count > 0 {
		return
	}

	users := []User{
		{Name: "张三", Email: "zhangsan@example.com"},
		{Name: "李四", Email: "lisi@example.com"},
	}
	db.Create(&users)

	posts := []Post{
		{Title: "Go语言入门", Content: "Go是一门简洁高效的编程语言...", UserID: users[0].ID},
		{Title: "GORM使用指南", Content: "GORM是Go语言中流行的ORM框架...", UserID: users[0].ID},
		{Title: "SQLite简介", Content: "SQLite是一个轻量级的嵌入式数据库...", UserID: users[1].ID},
	}
	db.Create(&posts)

	comments := []Comment{
		{Content: "写得真好！", PostID: posts[0].ID},
		{Content: "非常实用的教程", PostID: posts[0].ID},
		{Content: "学到了很多", PostID: posts[0].ID},
		{Content: "感谢分享", PostID: posts[1].ID},
		{Content: "SQLite确实很方便", PostID: posts[2].ID},
		{Content: "适合小项目使用", PostID: posts[2].ID},
	}
	db.Create(&comments)

	// 插入评论后更新文章评论状态
	db.Model(&Post{}).Where("id IN ?", []uint{posts[0].ID, posts[1].ID, posts[2].ID}).
		UpdateColumn("comment_status", "有评论")

	fmt.Println("=== 测试数据插入成功 ===")
}

// 查询某个用户发布的所有文章及其对应的评论
func queryUserPostsWithComments(db *gorm.DB) {
	fmt.Println("\n=== 题目2-1：查询用户【张三】的所有文章及评论 ===")

	var user User
	db.Where("name = ?", "张三").
		Preload("Posts.Comments").
		First(&user)

	fmt.Printf("用户: %s (Email: %s)\n", user.Name, user.Email)
	for _, post := range user.Posts {
		fmt.Printf("  文章: %s\n", post.Title)
		for _, comment := range post.Comments {
			fmt.Printf("    评论: %s\n", comment.Content)
		}
	}
}

// 查询评论数量最多的文章
func queryMostCommentedPost(db *gorm.DB) {
	fmt.Println("\n=== 题目2-2：查询评论数量最多的文章 ===")

	var result struct {
		PostID       uint
		CommentCount int64
	}
	db.Model(&Comment{}).
		Select("post_id, count(*) as comment_count").
		Group("post_id").
		Order("comment_count DESC").
		Limit(1).
		Scan(&result)

	var post Post
	db.Preload("Comments").First(&post, result.PostID)

	fmt.Printf("文章: %s (评论数: %d)\n", post.Title, result.CommentCount)
	for _, comment := range post.Comments {
		fmt.Printf("  评论: %s\n", comment.Content)
	}
}

// 演示钩子函数效果
func demonstrateHooks(db *gorm.DB) {
	fmt.Println("\n=== 题目3：钩子函数演示 ===")

	// 演示 Post AfterCreate 钩子：创建文章后自动更新用户文章数
	var user User
	db.Where("name = ?", "张三").First(&user)
	fmt.Printf("创建新文章前 - %s 的文章数: %d\n", user.Name, user.PostCount)

	newPost := Post{Title: "新文章：钩子函数详解", Content: "钩子函数可以在模型操作前后执行自定义逻辑...", UserID: user.ID}
	db.Create(&newPost)

	db.First(&user, user.ID)
	fmt.Printf("创建新文章后 - %s 的文章数: %d (AfterCreate钩子生效)\n", user.Name, user.PostCount)

	// 演示 Comment AfterDelete 钩子：删除文章的所有评论后状态变为"无评论"
	newComment := Comment{Content: "测试评论", PostID: newPost.ID}
	db.Create(&newComment)
	db.Model(&Post{}).Where("id = ?", newPost.ID).UpdateColumn("comment_status", "有评论")

	var postBefore Post
	db.First(&postBefore, newPost.ID)
	fmt.Printf("\n删除评论前 - 文章【%s】评论状态: %s\n", postBefore.Title, postBefore.CommentStatus)

	db.Delete(&newComment)

	var postAfter Post
	db.First(&postAfter, newPost.ID)
	fmt.Printf("删除评论后 - 文章【%s】评论状态: %s (AfterDelete钩子生效)\n", postAfter.Title, postAfter.CommentStatus)
}
