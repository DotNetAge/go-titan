package auth

import (
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	go_jwt "github.com/golang-jwt/jwt"
	uuid "github.com/satori/go.uuid"
)

// Tokens 访问令牌验证组件
type Tokens interface {
	Inspect(string) (*Principal, error)  // Inspect 检查 AccessToken是否有效，并从中读取用户身份信息
	Generate(*Principal) (*Token, error) // Generate 生成的令牌实例对象
	Options() interface{}                // Options 获取身份验证组件全部的选项
}

// Token 定义访问令牌的数据结构
type Token struct {
	AccessToken  string        `json:"access_token"`  // AccessToken 资源访问令牌
	RefreshToken string        `json:"refresh_token"` // RefreshToken 刷新令牌
	Created      time.Time     `json:"created"`       // Created 创建日期
	ExpiresIn    time.Duration `json:"expires_in"`    // ExpiresIn 过期时限
}

// Expired 判断令牌是否已经过期
func (t *Token) Expired() bool {
	return t.Created.Add(t.ExpiresIn).Unix() < time.Now().Unix()
}

// UserClaims JWT claims 仅供内部使用
type UserClaims struct {
	go_jwt.StandardClaims
	Principal
}

// NewUserClaims creates a new UserClaim
func NewUserClaims(user *Principal, duration time.Duration) *UserClaims {
	userClaims := &UserClaims{
		StandardClaims: go_jwt.StandardClaims{
			Issuer:    user.Issuer,                     // 签发者
			IssuedAt:  time.Now().Unix(),               // 签发时间
			ExpiresAt: time.Now().Add(duration).Unix(), // 过期时间
			Subject:   user.Name,                       // 签发给谁
		},
		Principal: *user,
	}
	return userClaims
}

// Principal 用于表示在微服务间传递的用户信息
type Principal struct {
	ID        string            `json:"id"`      // ID 用户ID
	Type      string            `json:"type"`    // Type 标识账号身份的类型, 例如: service
	Name      string            `json:"name"`    // Name 用户名
	Issuer    string            `json:"issuer"`  // Issuer 访问令牌的签发方
	IssuedAt  int64             `json:"issued"`  // IssuedAt 签发时间
	ExpiresAt int64             `json:"expires"` // ExpiresAt 过期时间
	Scopes    []string          `json:"scopes"`  // Scopes 私隐授权的范围
	Roles     []string          `json:"roles"`   // Roles 用户的角色
	Meta      map[string]string `json:"meta"`    // Meta 用户帐号中的其它信息
}

func NewUser(username string, scopes ...string) *Principal {
	user := &Principal{
		ID:   uuid.NewV4().String(),
		Meta: make(map[string]string),
	}
	if len(scopes) > 0 {
		user.Scopes = scopes
	}

	return user
}

// SetScopes 设置授权范围
func (user *Principal) SetScopes(scope ...string) *Principal {
	user.Scopes = scope
	return user
}

// SetRoles 设置用户角色
func (user *Principal) SetRoles(role ...string) *Principal {
	user.Roles = role
	return user
}

func (user *Principal) InRoles(roles ...string) bool {
	for _, r := range roles {
		if user.HasRole(r) {
			return true
		}
	}
	return false
}

func (user *Principal) HasRole(role string) bool {
	for _, n := range user.Roles {
		if n == role {
			return true
		}
	}

	return false
}

// Rules 提供基于授权范围的权限控制
type Rules interface {
	// Resources 加载系统定义的全部可访问资源
	Resources(domains ...string) ([]*Resource, error)
	// GetRes 获取指定ID的Resouce资源
	GetRes(id string) *Resource
	// Verify 检查用户的授权范围是否有效
	Verify(user *Principal, res *Resource) (bool, error)
	// VerifyRequest 验证请求是否合法
	VerifyRequest(req *http.Request) (bool, error)
}

// Resource 定义可访问资源
type Resource struct {
	ID       string `csv:"id" gorm:"primaryKey;type:varchar(200)"` // ID 标识资源的使用标识,即 scope 的定义
	Name     string `csv:"name" gorm:"type:varchar(200)"`          // Name 资源的中文使用说明
	Domain   string `csv:"domain" gorm:"type:varchar(1024)"`       // ClientID 用于标记当前资源所属的应用编号。也可以作为Domain使用
	Type     string `csv:"type" gorm:"type:varchar(50)"`           // Type 资源的类型 如 service、api、file
	Action   string `csv:"action" gorm:"type:varchar(50)"`         // Action HTTP方法POST、PUT、GET
	EndPoint string `csv:"url" gorm:"type:varchar(2048)"`          // EndPoint 资源指向的具体URL
}

// IdentityManager 用户身份标识管理器
type IdentityManager interface {
	// Valid 根据身份类型验证用户身份是否有效（默认仅支持内置三种方式）
	Valid(clientId, name, password string, t IdentityTypes) (*Identity, error)
	// GenVerifyCode 生成的随机手机验证码
	GenVerifyCode(clientId, mobile string) (string, error)
	// GetVerifyCode 获取手机验证码，读取后马上删除
	GetVerifyCode(clientId, mobile string) (string, error)
	// Add 创建新的身份标识
	Add(user *Identity) error
	// ChangePassword 更改用户密码
	ChangePassword(user *Identity, password string) error
	// Get 通过用户名获取身份标识
	Get(clientId, name string, idType IdentityTypes) (*Identity, error)
	// Bind 将身份标识绑定到指定的用户帐号上
	Bind(id, uid string) error
	// Remove 移除用户身份标识
	Remove(id string) error
	// List 列出指定用户ID下的全部身份标识
	List(clientId, userId string) ([]Identity, error)
	// Setup 初始化用户身分标识管理器的数据实体对象
	Setup() error
}

type IdentityTypes int16

const (
	IDUserNameAndPassword  IdentityTypes = 1
	IDEmailAndPassword     IdentityTypes = 2
	IDMobileAndVerifyCode  IdentityTypes = 3
	IDOAuth2NetworkAccount IdentityTypes = 4
	IDWeChatOpenID         IdentityTypes = 5
	IDWeChatUnionID        IdentityTypes = 6
)

// Identity 用户身份标识
type Identity struct {
	ID        string         `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" pg:"id,type:uuid,pk,default:uuid_generate_v4()"` // ID 身份标识ID
	ClientID  string         `json:"client_id" gorm:"type:varchar(50)" pg:"client_id,type:varchar(50)"`                                         // ClientID 客户端ID
	UserID    string         `json:"user_id" gorm:"type:varchar(50)" pg:"user_id,type:varchar(50)"`                                             // UserID 用户统一帐号ID。可空，但可通过Bind方法与用户账号进行绑定
	Type      IdentityTypes  `json:"id_type" gorm:"column:id_type",pg:"id_type,default:1"`                                                      // Type 身份标识的类型
	Name      string         `json:"name" gorm:"type:varchar(50)" pg:"name,type:varchar(50)"`                                                   // Name 身份标识的名称，可以为用户名、手机、或邮件
	Secret    []byte         `json:"secret" pg:"secret"`                                                                                        // Secret 用户机密信息，可以是密码
	MetaData  datatypes.JSON `json:"-" pg:",json_use_number"`                                                                                   // Meta 其它信息 TODO: 可以将存储的类型与使用的类型分离
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (user *Identity) SetPassword(password string) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user.Secret = hashedPassword
}

// IsCorrectPassword 检查输入密码是否正确
func (user *Identity) IsCorrectPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Secret), []byte(password))
	return err == nil
}

func NewEmailID(clientId, email string, password string) *Identity {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	return &Identity{
		ID:       uuid.NewV4().String(),
		ClientID: clientId,
		Type:     IDEmailAndPassword,
		Name:     email,
		Secret:   hashedPassword,
	}
}

func NewMobileID(clientId, mobile string) *Identity {
	return &Identity{
		ID:       uuid.NewV4().String(),
		ClientID: clientId,
		Type:     IDMobileAndVerifyCode,
		Name:     mobile,
	}
}

func NewOpenID(clientId, openID string) *Identity {
	return &Identity{
		ID:       uuid.NewV4().String(),
		ClientID: clientId,
		Type:     IDWeChatOpenID,
		Name:     openID,
	}
}

func NewUnionID(clientId, unionID string) *Identity {
	return &Identity{
		ID:       uuid.NewV4().String(),
		ClientID: clientId,
		Type:     IDWeChatUnionID,
		Name:     unionID,
	}
}

func NewUserPwdID(clientId, userName, password string) *Identity {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return &Identity{
		ID:       uuid.NewV4().String(),
		ClientID: clientId,
		Type:     IDUserNameAndPassword,
		Name:     userName,
		Secret:   hashedPassword,
	}
}
