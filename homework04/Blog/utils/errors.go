package utils

import "errors"

var (
	ErrUserNotFound       = errors.New("用户不存在")
	ErrInvalidCredentials = errors.New("用户名或密码错误")
	ErrPostNotFound       = errors.New("文章不存在")
	ErrEmptyTitle         = errors.New("文章标题不能为空")
	ErrEmptyComment       = errors.New("评论内容不能为空")
	ErrNoPermission       = errors.New("无权删除该文章")
	ErrEmptyFields        = errors.New("用户名、密码、邮箱不能为空")
	ErrEmailExists        = errors.New("该邮箱已被注册")
)
