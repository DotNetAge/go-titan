package jwt

import (
	"go-titan/auth"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestJWTAuth(t *testing.T) {
	duration := time.Minute

	man := New(
		SignWithHS256(),
		ExpiresIn(duration),
	)

	user := auth.NewUser("Ray", "user-info")

	token, err := man.Generate(user)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	principal, err := man.Inspect(token.AccessToken)
	require.NoError(t, err)
	require.NotEmpty(t, principal)

	require.Equal(t, user.ID, principal.ID)
}
