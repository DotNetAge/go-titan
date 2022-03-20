package paseto

import (
	"errors"
	"fmt"
	"go-titan/auth"
	"time"

	"github.com/aead/chacha20poly1305"
	go_paseto "github.com/o1egl/paseto"
)

type PasetoTokens struct {
	pto     *go_paseto.V2
	options Options
}

func New(opts ...Option) auth.Tokens {
	options := Options{}
	for _, opt := range opts {
		opt(&options)
	}

	if options.SymmetricKey == nil {
		options.SymmetricKey = []byte(auth.RandStr(32))
	}

	return &PasetoTokens{
		pto:     go_paseto.NewV2(),
		options: options,
	}
}

func (a *PasetoTokens) Options() interface{} {
	return a.options
}

//Generate 生成 AccessToken
func (p *PasetoTokens) Generate(user *auth.Principal) (*auth.Token, error) {
	symmetricKey := p.options.SymmetricKey.([]byte)
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("Key长度无效: 必须为%d个字母组成", chacha20poly1305.KeySize)
	}

	createAt := time.Now()
	user.IssuedAt = createAt.Unix()
	user.ExpiresAt = time.Now().Add(p.options.ExpiresIn).Unix()
	user.Issuer = p.options.Issuer

	accessToken, err := p.pto.Encrypt(symmetricKey, user, nil)
	if err != nil {
		return nil, err
	}
	token := &auth.Token{
		AccessToken: accessToken,
		Created:     createAt,
		ExpiresIn:   p.options.ExpiresIn,
	}

	user.ExpiresAt = time.Now().Add(p.options.RefreshExpiresIn).Unix()
	refreshToken, err := p.pto.Encrypt(symmetricKey, user, nil)
	if err != nil {
		return nil, err
	}
	token.RefreshToken = refreshToken
	return token, nil
}

// Verify 验证JWT的合法性并从中读取Principal实例
func (p *PasetoTokens) Inspect(token string) (*auth.Principal, error) {
	user := &auth.Principal{}
	err := p.pto.Decrypt(token, p.options.SymmetricKey.([]byte), user, nil)
	if err != nil {
		return nil, errors.New("无效的令牌")
	}
	return user, nil
}
