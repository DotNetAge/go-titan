package auth

import (
	"time"

	"github.com/dotnetage/go-titan/repository"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type idManager struct {
	repos       repository.Repository
	codeExpires time.Duration
}

func NewIDMan(repos repository.Repository) IdentityManager {
	return &idManager{
		repos:       repos,
		codeExpires: time.Second * 30,
	}
}

// Valid 根据身份类型验证用户身份是否有效（默认仅支持内置三种方式）
func (manager *idManager) Valid(clientId, name, password string, t IdentityTypes) (*Identity, error) {
	id, err := manager.Get(clientId, name, t)
	if err != nil {
		return nil, ErrUserNotFound
	}
	err = bcrypt.CompareHashAndPassword(id.Secret, []byte(password))
	if err != nil {
		return nil, ErrWrongPassword
	}
	return id, nil
}

// GenVerifyCode 生成的随机手机验证码
func (manager *idManager) GenVerifyCode(clientId, mobile string) (string, error) {
	c := NewVerifyCode(clientId, mobile, manager.codeExpires)
	exists, err := manager.repos.Exists(c)

	if err != nil {
		return "", err
	}

	if exists {
		if err := manager.repos.Update(c); err != nil {
			return "", err
		}
	} else {
		if err := manager.repos.Add(c); err != nil {
			return "", err
		}
	}

	return c.Code, nil
}

// GetVerifyCode 获取手机验证码，读取后马上删除
func (manager *idManager) GetVerifyCode(clientId, mobile string) (string, error) {
	c := &VerifyCode{
		Mobile:   mobile,
		ClientID: clientId,
	}

	if clientId == "" {
		return "", ErrClientIDNotFound
	}

	if mobile == "" {
		return "", ErrUserMobileNotFound
	}

	if err := manager.repos.Delete(c, "expire_at< ?", time.Now()); err != nil {
		return "", err
	}

	if err := manager.repos.Get(c); err != nil {
		return "", nil
	}

	return c.Code, nil
}

// Add 创建新的身份标识
func (manager *idManager) Add(user *Identity) error {
	if user.ID == "" {
		user.ID = uuid.NewV4().String()
	}
	return manager.repos.Add(user)
}

// Get 通过用户名获取身份标识
func (manager *idManager) Get(clientId,
	name string,
	idType IdentityTypes) (*Identity, error) {
	var user Identity
	_, err := manager.repos.Query(&user,
		"client_id = ? AND name = ? AND id_type = ?",
		clientId,
		name,
		idType)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Bind 将身份标识绑定到指定的用户帐号上
func (manager *idManager) Bind(id, uid string) error {
	return manager.repos.Update(&Identity{ID: id, UserID: uid})
}

// Remove 移除用户身份标识
func (manager *idManager) Remove(id string) error {
	return manager.repos.Delete(&Identity{ID: id})
}

// List 列出指定用户ID下的全部身份标识
func (manager *idManager) List(clientId, userId string) ([]Identity, error) {
	var ids []Identity
	_, err := manager.repos.Query(&ids,
		"client_id = ? and user_id = ?",
		clientId,
		userId)

	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (manager *idManager) Setup() error {
	return manager.repos.Setup(&Identity{}, &VerifyCode{})
}

func (manager *idManager) ChangePassword(user *Identity, password string) error {
	user.SetPassword(password)
	return manager.repos.Update(user)
}
