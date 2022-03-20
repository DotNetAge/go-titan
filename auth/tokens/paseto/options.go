package paseto

import (
	"time"

	"github.com/dotnetage/go-titan/auth"
)

type Option func(*Options)

type Options struct {
	Issuer string
	// SymmetricKey 对称密钥，专用于Paseto
	SymmetricKey interface{}

	ExpiresIn time.Duration

	RefreshExpiresIn time.Duration
}

func SignWithPaseto(key ...interface{}) Option {
	return func(o *Options) {
		if len(key) == 0 {
			o.SymmetricKey = []byte(auth.RandStr(32))
		} else {
			o.SymmetricKey = key
		}
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
