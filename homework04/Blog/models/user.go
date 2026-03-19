package models

import "gorm.io/gorm"

type User struct {
	Email string `gorm:"size:200;uniqueIndex;not null"`
	gorm.Model
	Username  string `gorm:"size:100;not null"`
	Password  string `gorm:"size:100;not null" json:"-"`
	PostCount int    `gorm:"default:0"`
	PostIDs   []Post `gorm:"foreignKey:UserID"`
}

type Post struct {
	gorm.Model
	Title         string    `gorm:"size:200;not null"`
	Content       string    `gorm:"type:text"`
	UserID        uint      `gorm:"not null;index"`
	Author        *User     `gorm:"foreignKey:UserID" json:"author,omitempty"`
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
	Author    *User  `gorm:"foreignKey:UserID" json:"author,omitempty"`
	userName  string
	postTitle string
}

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

func (p *Post) AfterCreate(tx *gorm.DB) (err error) {
	return tx.Model(&User{}).Where("id = ?", p.UserID).
		UpdateColumn("PostCount", gorm.Expr("PostCount + 1")).Error
}

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

func (c *Comment) AfterCreate(tx *gorm.DB) (err error) {
	return tx.Model(&Post{}).Where("id = ?", c.PostID).
		UpdateColumns(map[string]interface{}{
			"CommentCount":  gorm.Expr("CommentCount + 1"),
			"CommentStatus": "有评论",
		}).Error
}

func (c *Comment) AfterDelete(tx *gorm.DB) (err error) {
	var count int64
	if err := tx.Model(&Comment{}).Where("PostID = ?", c.PostID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return tx.Model(&Post{}).Where("id = ?", c.PostID).
			UpdateColumns(map[string]interface{}{
				"CommentStatus": "无评论",
				"CommentCount":  0,
			}).Error
	}
	return tx.Model(&Post{}).Where("id = ?", c.PostID).
		UpdateColumn("CommentCount", gorm.Expr("CommentCount - 1")).Error
}
