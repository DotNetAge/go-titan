package auth

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/dotnetage/go-titan/repository"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func newRepos() repository.Repository {
	repos := repository.New(repository.WithPostgre(
		repository.Database("auth_db"),
		repository.UserName("titan"),
		repository.Password("titan"),
	))
	return repos
}

var (
	repos repository.Repository
	ids   IdentityManager
)

func init() {
	repos := newRepos()
	ids := NewIDMan(repos)
	ids.Setup()
}

func TestVerifyCode(t *testing.T) {

	clientId := uuid.NewV4().String()
	mobile := gofakeit.Phone()

	code, err := ids.GenVerifyCode(clientId, mobile)
	require.NoError(t, err)
	require.NotEmpty(t, code)

	code2, err := ids.GetVerifyCode(clientId, mobile)
	require.NoError(t, err)
	require.NotEmpty(t, code2)
	require.Equal(t, code, code2)
}

func TestIdentityValid(t *testing.T) {
	// repos := newRepos()
	// ids := New(repos)
	// ids.Setup()
	clientId := gofakeit.UUID()
	pwd := gofakeit.Password(true, true, true, true, false, 8)
	pwd1 := gofakeit.Password(true, true, true, true, false, 8)
	emailID := NewEmailID(clientId, gofakeit.Email(), pwd)
	userPwdID := NewUserPwdID(clientId, gofakeit.Username(), pwd1)
	err := ids.Add(emailID)
	require.NoError(t, err)
	err = ids.Add(userPwdID)
	require.NoError(t, err)

	eid, err := ids.Valid(clientId, emailID.Name, pwd, IDEmailAndPassword)
	require.NoError(t, err)
	require.Equal(t, emailID.ID, eid.ID)

	uid, err := ids.Valid(clientId, userPwdID.Name, pwd1, IDUserNameAndPassword)
	require.NoError(t, err)
	require.Equal(t, uid.ID, userPwdID.ID)

	err = ids.ChangePassword(uid, pwd)
	require.NoError(t, err)

	err = ids.Remove(uid.ID)
	require.NoError(t, err)
}
