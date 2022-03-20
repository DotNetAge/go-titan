package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRules(t *testing.T) {
	res := []Resource{
		{
			ID:       "accounts",
			Name:     "帐号管理",
			Domain:   "www.titan.com",
			Type:     "*",
			EndPoint: "/v1/api/accounts/*",
		},
		{
			ID:       "accounts:read",
			Name:     "生成读取用户信息的权限",
			Domain:   "www.titan.com",
			Type:     "get",
			EndPoint: "/v1/api/accounts/{id}",
		},
		{
			ID:       "accounts:write",
			Name:     "生成可更新用户信息的权限",
			Domain:   "www.titan.com",
			Type:     "put",
			EndPoint: "/v1/api/accounts/{id}",
		},
	}
	// repos := repository.New(repository.WithSQLite("test.db"))

	// rules, err := NewRules("res_data.csv")
	// require.NoError(t, err)
	// user := &Principal{
	// 	Scopes: []string{
	// 		"*",
	// 	}}

	// ok, err := rules.Verify(user, &res[2])
	// require.NoError(t, err)
	// require.Equal(t, ok, true)
	require.NotNil(t, res)
}
