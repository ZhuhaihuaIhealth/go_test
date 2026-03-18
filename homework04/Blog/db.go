package main

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// user 表
type User struct {
	Email string `gorm:"size:200;uniqueIndex;not null"`
	gorm.Model
	Username  string `gorm:"size:100;not null"`
	Password  string `gorm:"size:100;not null" json:"-"`
	PostCount int    `gorm:"default:0"`
	PostIDs   []Post `gorm:"foreignKey:UserID"`
}

// post 表
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

// comment 表
type Comment struct {
	gorm.Model
	PostID    uint   `gorm:"not null;index"`
	Content   string `gorm:"type:text;not null"`
	UserID    uint   `gorm:"not null;index"`
	userName  string
	postTitle string
}

// 文章创建前钩子函数，若未指定 UserID 则根据用户名查找
func (p *Post) BeforeCreate(tx *gorm.DB) (err error) {
	if p.UserID != 0 {
		return nil
	}
	var user User
	if err := tx.Where("Username = ?", p.userName).First(&user).Error; err != nil {
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

// 创建Comment前的钩子函数，若未指定 PostID/UserID 则根据名称查找
func (c *Comment) BeforeCreate(tx *gorm.DB) (err error) {
	if c.PostID != 0 && c.UserID != 0 {
		return nil
	}
	if c.PostID == 0 {
		var post Post
		if err := tx.Where("Title = ?", c.postTitle).First(&post).Error; err != nil {
			return err
		}
		c.PostID = post.ID
	}
	if c.UserID == 0 {
		var user User
		if err := tx.Where("Username = ?", c.userName).First(&user).Error; err != nil {
			return err
		}
		c.UserID = user.ID
	}
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
	return tx.Model(&Post{}).Where("id = ?", c.PostID).UpdateColumns(map[string]interface{}{"CommentCount": gorm.Expr("CommentCount + 1"), "CommentStatus": "有评论"}).Error
}
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

// GetAllPosts 获取所有用户的文章（含作者用户名）
func GetAllPosts(db *gorm.DB) ([]map[string]interface{}, error) {
	var posts []Post
	if err := db.Order("CreatedAt DESC").Find(&posts).Error; err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	for _, p := range posts {
		var user User
		username := "未知用户"
		if err := db.First(&user, p.UserID).Error; err == nil {
			username = user.Username
		}
		result = append(result, map[string]interface{}{
			"ID":        p.ID,
			"Title":     p.Title,
			"Content":   p.Content,
			"UserID":    p.UserID,
			"Username":  username,
			"CreatedAt": p.CreatedAt,
		})
	}
	return result, nil
}

// GetUserPosts 获取指定用户的所有文章
func GetUserPosts(db *gorm.DB, userID uint) ([]Post, error) {
	var posts []Post
	if err := db.Where("UserID = ?", userID).Order("CreatedAt DESC").Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

// CreateUserPost 为指定用户创建文章
func CreateUserPost(db *gorm.DB, userID uint, title, content string) (*Post, error) {
	if title == "" {
		return nil, errors.New("文章标题不能为空")
	}
	post := Post{
		Title:   title,
		Content: content,
		UserID:  userID,
	}
	if err := db.Create(&post).Error; err != nil {
		return nil, errors.New("文章创建失败: " + err.Error())
	}
	return &post, nil
}

// DeleteUserPost 删除指定用户的文章（只能删除自己的）
func DeleteUserPost(db *gorm.DB, userID uint, postID uint) error {
	var post Post
	if err := db.First(&post, postID).Error; err != nil {
		return errors.New("文章不存在")
	}
	if post.UserID != userID {
		return errors.New("无权删除该文章")
	}
	if err := db.Where("PostID = ?", postID).Delete(&Comment{}).Error; err != nil {
		return errors.New("删除文章评论失败: " + err.Error())
	}
	if err := db.Delete(&post).Error; err != nil {
		return errors.New("文章删除失败: " + err.Error())
	}
	return nil
}

// GetPostDetail 获取文章详情及其所有评论（含评论者用户名）
func GetPostDetail(db *gorm.DB, postID uint) (*Post, []map[string]interface{}, error) {
	var post Post
	if err := db.First(&post, postID).Error; err != nil {
		return nil, nil, errors.New("文章不存在")
	}
	var comments []Comment
	db.Where("PostID = ?", postID).Order("CreatedAt ASC").Find(&comments)

	var result []map[string]interface{}
	for _, c := range comments {
		username := "未知用户"
		var user User
		if err := db.First(&user, c.UserID).Error; err == nil {
			username = user.Username
		}
		result = append(result, map[string]interface{}{
			"ID":        c.ID,
			"PostID":    c.PostID,
			"UserID":    c.UserID,
			"Username":  username,
			"Content":   c.Content,
			"CreatedAt": c.CreatedAt,
		})
	}
	return &post, result, nil
}

// CreateComment 为指定文章创建评论
func CreateComment(db *gorm.DB, userID uint, postID uint, content string) (*Comment, error) {
	if content == "" {
		return nil, errors.New("评论内容不能为空")
	}
	var post Post
	if err := db.First(&post, postID).Error; err != nil {
		return nil, errors.New("文章不存在")
	}
	comment := Comment{
		PostID:  postID,
		UserID:  userID,
		Content: content,
	}
	if err := db.Create(&comment).Error; err != nil {
		return nil, errors.New("评论创建失败: " + err.Error())
	}
	return &comment, nil
}

// GetUser 根据用户名从数据库获取用户信息
func GetUser(db *gorm.DB, username string) (*User, error) {
	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	return &user, nil
}

// RegisterUser 注册用户，密码使用 bcrypt 加密存储
func RegisterUser(db *gorm.DB, username, password, email string) (*User, error) {
	if username == "" || password == "" || email == "" {
		return nil, errors.New("用户名、密码、邮箱不能为空")
	}

	var existing User
	if err := db.Where("email = ?", email).First(&existing).Error; err == nil {
		return nil, errors.New("该邮箱已被注册")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败: " + err.Error())
	}

	user := User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
	}
	if err := db.Create(&user).Error; err != nil {
		return nil, errors.New("用户创建失败: " + err.Error())
	}
	return &user, nil
}
