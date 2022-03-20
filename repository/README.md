# Repository 模式

```go 
import (
    "go-titan/repository"
)

func main() {
    repos := repository.New(
        repository.WithPostgre(
            DBName("testdb"),
            DBUser("titan"),
            DBPwd("titan"),
        ))
}
```

## 建模

titan 的默认使用gorm实现`Repository`，模型的声明请参考 GORM的[声明模型](https://gorm.io/zh_CN/docs/models.html)中的内容

由于是通用模型所以对特殊的字段类型建议使用`gorm.io/datatypes`提供的数据类型。

```go
type ComposeData struct {
	ID     string `gorm:"primaryKey;"`
	Secret []byte
	Tags  []string `gorm:"type:text[]"`
	Metas datatypes.JSON
}
```

## 新增

```go
testUser := &UserTest{
		Name:  gofakeit.UserName(),
		Email: gofakeit.Email(),
	}
err = repos.Add(testUser)
require.NoError(t, err)
```

## 删除

## 查询

## 更新