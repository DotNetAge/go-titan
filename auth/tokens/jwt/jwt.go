package jwt

import (
	"errors"
	"fmt"
	"go-titan/auth"
	"sync"
	"time"

	go_jwt "github.com/golang-jwt/jwt"
)

type JWTokens struct {
	sync.Mutex
	options Options
}

func New(opts ...Option) auth.Tokens {
	options := Options{
		Issuer:           "www.titan.com",
		ExpiresIn:        time.Second * 7200,  // 2小时
		RefreshExpiresIn: time.Hour * 24 * 25, // 25 天过期
	}

	for _, opt := range opts {
		opt(&options)
	}

	return &JWTokens{options: options}
}

func (a *JWTokens) Options() interface{} {
	a.Lock()
	defer a.Unlock()
	return a.options
}

func (a *JWTokens) Generate(user *auth.Principal) (*auth.Token, error) {
	createdAt := time.Now()
	userClaims := &auth.UserClaims{
		StandardClaims: go_jwt.StandardClaims{
			Issuer:    a.options.Issuer,                           // 签发者
			IssuedAt:  createdAt.Unix(),                           // 签发时间
			ExpiresAt: time.Now().Add(a.options.ExpiresIn).Unix(), // 过期时间
			Subject:   user.Name,                                  // 签发给谁
		},
		Principal: *user,
	}
	acc := go_jwt.NewWithClaims(a.options.SigningMethod, userClaims)
	accessToken, err := acc.SignedString(a.options.PrivateKey)
	if err != nil {
		return nil, err
	}
	userClaims.StandardClaims.ExpiresAt = time.Now().Add(a.options.RefreshExpiresIn).Unix()
	ref := go_jwt.NewWithClaims(a.options.SigningMethod, userClaims)
	refreshToken, err := ref.SignedString(a.options.PrivateKey)
	if err != nil {
		return nil, err
	}
	return &auth.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Created:      createdAt,
		ExpiresIn:    a.options.ExpiresIn,
	}, nil
}

func (ja *JWTokens) Inspect(accessToken string) (*auth.Principal, error) {
	token, err := go_jwt.ParseWithClaims(
		accessToken,
		&auth.UserClaims{},
		func(token *go_jwt.Token) (interface{}, error) {
			if ja.options.SigningMethod != token.Method {
				return nil, errors.New("未能识别的签名方法")
			}
			return ja.options.PublicKey, nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("非法的令牌: %w", err)
	}

	claims, ok := token.Claims.(*auth.UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return &claims.Principal, nil
}
