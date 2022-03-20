package auth

import (
	"context"
	"errors"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"
)

var (
	runeMask = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

type (
	CurrentUserKey struct{}
	XClientKey     struct{}
)

func GenVerifyCode() string {
	return RandStrNumber(4, []rune("1234567890"))
}

func GenAuthCode() string {
	return RandStrNumber(8, runeMask)
}

// RandStr 生成指定长度的随机字符串
func RandStr(n int) string {
	return RandStrNumber(n, runeMask)
}

func RandStrNumber(n int, mask []rune) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = mask[rand.Intn(len(mask))]
	}
	return string(b)
}

// AuthUser 获取当前已进行授权验证的用户对象，如果失败则返回false
func AuthUser(ctx context.Context) (*Principal, bool) {
	acc, ok := ctx.Value(CurrentUserKey{}).(*Principal)
	return acc, ok
}

// AuthClient 获取当前请求中的ClientID
func AuthClient(ctx context.Context) (string, bool) {
	clientID, ok := ctx.Value(XClientKey{}).(string)
	return clientID, ok
}

func LoadKeyFile(keyFile string) ([]byte, error) {
	if len(keyFile) == 0 {
		return nil, errors.New("密钥文件不能为空")
	}
	pkFile, err := os.Open(keyFile)
	defer pkFile.Close()

	if err != nil {
		return nil, err
	}

	pkBytes, err := ioutil.ReadAll(pkFile)

	if err != nil {
		return nil, err
	}

	return pkBytes, nil
}

// ContextWithUser 将用户对象写入当前安全上下文中
func ContextWithUser(ctx context.Context, account *Principal) context.Context {
	return context.WithValue(ctx, CurrentUserKey{}, account)
}

func ContextWithClient(ctx context.Context, client string) context.Context {
	return context.WithValue(ctx, XClientKey{}, client)
}

func IsPatternMatch(key1 string, key2 string) bool {
	i := strings.Index(key2, "*")
	if i == -1 {
		return key1 == key2
	}

	if len(key1) > i {
		return key1[:i] == key2[:i]
	}
	return key1 == key2[:i]
}
