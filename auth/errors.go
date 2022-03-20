package auth

import "errors"

var (
	ErrWrongPassword      = errors.New("用户密码错误")
	ErrClientIDNotFound   = errors.New("缺少授权客户端ID")
	ErrUserNotFound       = errors.New("用户不存在")
	ErrUserMobileNotFound = errors.New("缺少获取验证码的手机号")
)
