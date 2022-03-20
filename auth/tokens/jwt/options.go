package jwt

import (
	"go-titan/auth"
	"time"

	go_jwt "github.com/golang-jwt/jwt"
)

type Option func(*Options)

type Options struct {
	Issuer string
	// SignMethod JWT的签发方法
	SigningMethod go_jwt.SigningMethod
	// PrivateKey 签名时使用的密钥 （RSA专用）
	PrivateKey interface{}
	// PublicKey 验证与读取JWT时使用的公钥（RSA专用）
	PublicKey interface{}

	// SigningKey 通用密钥
	SigningKey interface{}

	ExpiresIn time.Duration

	RefreshExpiresIn time.Duration
}

func SigningWith(m go_jwt.SigningMethod) Option {
	return func(o *Options) {
		o.SigningMethod = m
	}
}

func SigningKey(key interface{}) Option {
	return func(o *Options) {
		o.SigningKey = key
	}
}

func PrivateKey(privateKeyFile string) Option {
	return func(o *Options) {
		privateKey, _ := auth.LoadKeyFile(privateKeyFile)
		o.SigningMethod = go_jwt.SigningMethodRS256
		pKey, err := go_jwt.ParseRSAPrivateKeyFromPEM(privateKey)
		if err == nil {
			o.PrivateKey = pKey
		}
	}
}

func PublicKey(publicKeyFile string) Option {
	return func(o *Options) {
		publicKey, _ := auth.LoadKeyFile(publicKeyFile)
		o.SigningMethod = go_jwt.SigningMethodRS256
		pKey, err := go_jwt.ParseRSAPublicKeyFromPEM(publicKey)
		if err == nil {
			o.PublicKey = pKey
		}
	}
}

func SignWithHS256(key ...interface{}) Option {
	return func(o *Options) {
		o.SigningMethod = go_jwt.SigningMethodHS256
		if len(key) == 0 {
			o.PrivateKey = []byte(auth.RandStr(32))
		} else {
			o.PrivateKey = key
		}
		o.PublicKey = o.PrivateKey
	}

}

func Issuer(s string) Option {
	return func(o *Options) {
		o.Issuer = s
	}
}

func ExpiresIn(exp time.Duration) Option {
	return func(o *Options) {
		o.ExpiresIn = exp
	}
}

func RefreshExpiresIn(exp time.Duration) Option {
	return func(o *Options) {
		o.RefreshExpiresIn = exp
	}
}
