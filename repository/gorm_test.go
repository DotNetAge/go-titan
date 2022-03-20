package repository

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"
)

type UserTest struct {
	gorm.Model
	Name  string
	Email string
}

type ComposeData struct {
	ID     string `gorm:"primaryKey;"`
	Secret []byte
	Tags   []string `gorm:"type:text[]"`
	Metas  datatypes.JSON
}

func TestRepos(t *testing.T) {
	//dbFile := "test.db"
	//if _, err := os.Stat(dbFile); err == nil {
	//	if err := os.Remove(dbFile); err != nil {
	//		panic(err)
	//	}
	//}
	//db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})

	// dsn := "host=localhost user=titan password=titan dbname=auth_db port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	// db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	db := WithPostgre(
		Database("auth_db"),
		UserName("titan"),
		Password("titan"),
	)

	// require.NoError(t, err)
	entityType := &UserTest{}
	err := db.AutoMigrate(entityType)
	repos := New(db)
	userName := gofakeit.Name()
	testUser := &UserTest{
		Name:  userName,
		Email: gofakeit.Email(),
	}
	err = repos.Add(testUser)
	require.NoError(t, err)
	count, err := repos.Count(entityType)
	require.NoError(t, err)
	require.Equal(t, count, 1)
	require.Greater(t, testUser.ID, uint(0))

	testUser.Name = gofakeit.Username()
	err = repos.Update(testUser)
	require.NoError(t, err)

	testUser2 := &UserTest{
		Model: gorm.Model{
			ID: testUser.ID,
		},
	}

	err = repos.Get(testUser2)
	require.NoError(t, err)
	require.Equal(t, testUser.Name, testUser2.Name)

	err = repos.Delete(testUser)
	require.NoError(t, err)

}

func TestPGDataType(t *testing.T) {
	dsn := "host=localhost user=titan password=titan dbname=auth_db port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	entityType := &ComposeData{}

	_ = db.Migrator().DropTable(entityType)
	err = db.AutoMigrate(entityType)
	require.NoError(t, err)
}
