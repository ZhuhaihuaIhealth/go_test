package services

import (
	"BlogSystem/models"
	"BlogSystem/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// dummyHash is a pre-computed bcrypt hash used to prevent timing attacks.
// When a user is not found, we still run bcrypt.CompareHashAndPassword against
// this dummy so the response time is indistinguishable from a wrong-password case.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-timing-defense"), bcrypt.DefaultCost)

func Login(db *gorm.DB, username, password string) (string, error) {
	user, err := GetUser(db, username)
	if err != nil {
		bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return "", utils.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", utils.ErrInvalidCredentials
	}
	return utils.GenerateToken(user.ID, username)
}

func GetAllPosts(db *gorm.DB) ([]map[string]interface{}, error) {
	var posts []models.Post
	if err := db.Preload("Author").Order("CreatedAt DESC").Find(&posts).Error; err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	for _, p := range posts {
		username := "未知用户"
		if p.Author != nil {
			username = p.Author.Username
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

func GetUserPosts(db *gorm.DB, userID uint) ([]models.Post, error) {
	var posts []models.Post
	if err := db.Where("UserID = ?", userID).Order("CreatedAt DESC").Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func CreateUserPost(db *gorm.DB, userID uint, title, content string) (*models.Post, error) {
	if title == "" {
		return nil, utils.ErrEmptyTitle
	}
	post := models.Post{
		Title:   title,
		Content: content,
		UserID:  userID,
	}
	if err := db.Create(&post).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func DeleteUserPost(db *gorm.DB, userID uint, postID uint) error {
	var post models.Post
	if err := db.First(&post, postID).Error; err != nil {
		return utils.ErrPostNotFound
	}
	if post.UserID != userID {
		return utils.ErrNoPermission
	}
	if err := db.Where("PostID = ?", postID).Delete(&models.Comment{}).Error; err != nil {
		return err
	}
	if err := db.Delete(&post).Error; err != nil {
		return err
	}
	return nil
}

func GetPostDetail(db *gorm.DB, postID uint) (*models.Post, []map[string]interface{}, error) {
	var post models.Post
	if err := db.First(&post, postID).Error; err != nil {
		return nil, nil, utils.ErrPostNotFound
	}
	var comments []models.Comment
	db.Preload("Author").Where("PostID = ?", postID).Order("CreatedAt ASC").Find(&comments)

	var result []map[string]interface{}
	for _, c := range comments {
		username := "未知用户"
		if c.Author != nil {
			username = c.Author.Username
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

func CreateComment(db *gorm.DB, userID uint, postID uint, content string) (*models.Comment, error) {
	if content == "" {
		return nil, utils.ErrEmptyComment
	}
	var post models.Post
	if err := db.First(&post, postID).Error; err != nil {
		return nil, utils.ErrPostNotFound
	}
	comment := models.Comment{
		PostID:  postID,
		UserID:  userID,
		Content: content,
	}
	if err := db.Create(&comment).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func GetUser(db *gorm.DB, username string) (*models.User, error) {
	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, utils.ErrUserNotFound
	}
	return &user, nil
}

func RegisterUser(db *gorm.DB, username, password, email string) (*models.User, error) {
	if username == "" || password == "" || email == "" {
		return nil, utils.ErrEmptyFields
	}

	var existing models.User
	if err := db.Where("email = ?", email).First(&existing).Error; err == nil {
		return nil, utils.ErrEmailExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
	}
	if err := db.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
