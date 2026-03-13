package homeworkA

import (
	"testing"

	"gorm.io/gorm"
	dbTestutil "homework03/testutil"
)

type User struct {
	gorm.Model
	Email     string `gorm:"size:200;uniqueIndex;not null"`
	Name      string `gorm:"size:100;not null"`
	PostCount int    `gorm:"default:0"`
	PostIDs   []Post `gorm:"foreignKey:UserID"`
}
type Post struct {
	gorm.Model
	Title         string    `gorm:"size:200;not null"`
	Content       string    `gorm:"type:text"`
	UserID        uint      `gorm:"not null;index"`
	CommentCount  int       `gorm:"default:0"`
	CommentIDs    []Comment `gorm:"foreignKey:PostID"`
	CommentStatus string    `gorm:"size:50;default:无评论"`
	userName      string
}
type Comment struct {
	gorm.Model
	PostID    uint   `gorm:"not null;index"`
	Content   string `gorm:"type:text;not null"`
	UserID    uint   `gorm:"not null;index"`
	userName  string
	postTitle string
}

// 题目3 钩子函数
// 文章创建前钩子函数，根据用户名获取用户ID
func (p *Post) BeforeCreate(tx *gorm.DB) (err error) {
	var user User
	if err := tx.Where("name = ?", p.userName).First(&user).Error; err != nil {
		return err
	}
	p.UserID = user.ID
	return nil
}

// 创建Post时自动更新用户的PostCount字段
func (p *Post) AfterCreate(tx *gorm.DB) (err error) {
	var user User
	if err := tx.First(&user, p.UserID).Error; err != nil {
		return err
	}
	user.PostCount++
	//将文章的id添加到用户的PostIDs字段中
	return tx.Model(&User{}).Where("id = ?", p.UserID).
		UpdateColumn("PostCount", gorm.Expr("PostCount + 1")).Error
}

// 创建Comment前的钩子函数，根据文章名和用户名获取文章id和用户id
func (c *Comment) BeforeCreate(tx *gorm.DB) (err error) {
	var post Post
	if err := tx.Where("title = ?", c.postTitle).First(&post).Error; err != nil {
		return err
	}
	c.PostID = post.ID

	var user User
	if err := tx.Where("name = ?", c.userName).First(&user).Error; err != nil {
		return err
	}
	c.UserID = user.ID
	return nil
}

// 创建Comment后自动更新Post的CommentCount字段
func (c *Comment) AfterCreate(tx *gorm.DB) (err error) {
	/*var post Post
	if err := tx.First(&post, c.PostID).Error; err != nil {
		return err
	}
	post.CommentCount++
	if err := tx.Save(&post).Error; err != nil {
		return err
	}
	return nil*/
	return tx.Model(&Post{}).Where("id = ?", c.PostID).UpdateColumns(map[string]interface{}{"CommentCount": gorm.Expr("CommentCount + 1"),"CommentStatus": "有评论"}).Error
}

// 删除评论时检查并修改Post的CommentCount字段，当评论数减为0时，将Post的CommentStatus字段设置为无评论
func (c *Comment) AfterDelete(tx *gorm.DB) (err error) {
	var count int64
	if err := tx.Model(&Comment{}).Where("PostID = ?", c.PostID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		// 修改Post的CommentCount字段
		return tx.Model(&Post{}).Where("id = ?", c.PostID).
			UpdateColumns(map[string]interface{}{
				"CommentStatus": "无评论",
            	"CommentCount":  0,
			}).Error
	} else {
		return tx.Model(&Post{}).Where("id = ?", c.PostID).
			UpdateColumn("CommentCount", gorm.Expr("CommentCount - 1")).Error
	}
}
func TestBlog(t *testing.T) {
	db := dbTestutil.NewTestDB(t, "blog_sqlite.db")

	if err := db.AutoMigrate(&User{}, &Post{}, &Comment{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}
	users := []User{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
		{Name: "Celine", Email: "celine@example.com"},
	}
	if err := db.CreateInBatches(users, 3).Error; err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}
	// 创建帖子
	posts := []Post{
		{Title: "Post 1", Content: "This is the first post.", userName: "Alice"},
		{Title: "Post 2", Content: "This is the second post.", userName: "Alice"},
		{Title: "Post 3", Content: "This is the third post.", userName: "Celine"},
		{Title: "Post 4", Content: "This is the four post.", userName: "Alice"},
	}
	if err := db.CreateInBatches(posts, 3).Error; err != nil {
		t.Fatalf("Failed to create posts: %v", err)
	}
	//查询Alice
	var alice User
	if err := db.Where("name = ?", "Alice").First(&alice).Error; err != nil {
		t.Fatalf("Failed to query user: %v", err)
	}
	//打印查询到的内容
	t.Logf("User: %+v", alice)

	//创建评论
	comments := []Comment{
		{Content: "This is the first comment.", userName: "Alice", postTitle: "Post 1"},
		{Content: "This is the second comment.", userName: "Bob", postTitle: "Post 1"},
		{Content: "This is the third comment.", userName: "Celine", postTitle: "Post 2"},
		{Content: "This is the four comment.", userName: "Celine", postTitle: "Post 1"},
		{Content: "This is the five comment.", userName: "Celine", postTitle: "Post 4"},
	}
	if err := db.CreateInBatches(comments, 3).Error; err != nil {
		t.Fatalf("Failed to create comments: %v", err)
	}
	//查询Post 1的内容
	var post1 Post
	if err := db.Where("title = ?", "Post 1").First(&post1).Error; err != nil {
		t.Fatalf("Failed to query Post 1: %v", err)
	}
	//打印查询到的内容
	t.Logf("Post 1: %+v", post1)

	//Post 1删除一个评论
	var comment Comment
	if err := db.Where("content = ?", "This is the first comment.").First(&comment).Error; err != nil {
		t.Fatalf("Failed to query comment: %v", err)
	}
	if err := db.Delete(&comment).Error; err != nil {
		t.Fatalf("Failed to delete comment: %v", err)
	}
	//查询Post 1的内容
	if err := db.Preload("CommentIDs").First(&post1).Error; err != nil {
		t.Fatalf("Failed to query Post 1: %v", err)
	}
	t.Logf("Post 1: %+v", post1)
	//Post 2删除一个评论
	var comment2 Comment
	if err := db.Where("content = ?", "This is the third comment.").First(&comment2).Error; err != nil {
		t.Fatalf("Failed to query comment: %v", err)
	}
	if err := db.Delete(&comment2).Error; err != nil {
		t.Fatalf("Failed to delete comment: %v", err)
	}
	//查询Post 2的内容
	var post2 Post
	if err := db.Preload("CommentIDs").Where("title = ?", "Post 2").First(&post2).Error; err != nil {
		t.Fatalf("Failed to query Post 2: %v", err)
	}
	t.Logf("Post 2: %+v", post2)

	//题目二 
	//使用GORM查询某个用户发布的所有文章及其对应的评论信息
	var user User
	if err := db.Where("name = ?", "Alice").Preload("PostIDs.CommentIDs").First(&user).Error; err != nil {
		t.Fatalf("Failed to query user: %v", err)
	}
	// t.Logf("某个User的所有文章及评论信息: %+v", user)
	for _, post := range user.PostIDs {
		t.Logf("文章: %s", post.Title)
		for _, comment := range post.CommentIDs {
			t.Logf("评论: %s", comment.Content)
		}
	}

	//使用GORM查询评论数量最多的文章信息
	var mostCommentedPost Post
	if err := db.Order("CommentCount desc").First(&mostCommentedPost).Error; err != nil {
		t.Fatalf("Failed to query most commented post: %v", err)
	}
	t.Logf("评论数量最多的文章: %+v", mostCommentedPost)
}
