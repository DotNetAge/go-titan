package paseto

import (
	"testing"

	"github.com/dotnetage/go-titan/auth"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func TestPasetoManager(t *testing.T) {
	pasetoAuth := New()

	user := &auth.Principal{
		ID:     uuid.NewV4().String(),
		Name:   "Ray",
		Scopes: []string{"user-info"},
		Roles:  []string{"admin"},
	}

	token, err := pasetoAuth.Generate(user)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := pasetoAuth.Inspect(token.AccessToken)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	require.Equal(t, user.ID, payload.ID)

	payload, err = pasetoAuth.Inspect(token.RefreshToken)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.Equal(t, user.ID, payload.ID)

}
