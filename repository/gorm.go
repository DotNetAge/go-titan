package repository

import "gorm.io/gorm"

type gormRepos struct {
	// options *Options
	db    *gorm.DB
	limit int
	skip  int
}

func New(db *gorm.DB) Repository {
	//options := &Options{}
	//for _, opt := range opts {
	//	opt(options)
	//}
	//return &gormRepos{options: options}
	return &gormRepos{db: db}
}

func (r *gormRepos) Limit(limit int) Repository {
	r.limit = limit
	return r
}

func (r *gormRepos) Skip(skip int) Repository {
	r.skip = skip
	return r
}

func (r *gormRepos) Add(model interface{}) error {
	if model == nil {
		return ErrEntityCannotBeNull
	}

	result := r.db.Create(model)
	return result.Error

}

// Update 更新数据对象
func (r *gormRepos) Update(model interface{}) error {
	result := r.db.Updates(model)
	return result.Error
}

// Delete 删除指定的数据对象
func (r *gormRepos) Delete(model interface{}, params ...interface{}) error {
	result := r.db.Delete(model, params)
	return result.Error
}

func (r *gormRepos) Get(model interface{}, params ...interface{}) error {
	if len(params) == 0 {
		result := r.db.First(model)
		return result.Error
	} else {
		result := r.db.First(model, params...)
		return result.Error
	}
}

func (r *gormRepos) All(model interface{}) (int, error) {
	count := int64(0)
	result := r.db.Limit(r.limit).Offset(r.skip).Find(model)
	r.limit = 0
	r.skip = 0
	cntResult := r.db.Model(model).Count(&count)
	if cntResult.Error != nil {
		return 0, result.Error
	}
	return int(count), result.Error
}

func (r *gormRepos) Query(model interface{}, condition string, params ...interface{}) (int, error) {
	result := r.db.Where(condition, params...).
		Limit(r.limit).
		Offset(r.skip).
		Find(model)
	count := int64(0)

	cntResult := r.db.Model(model).
		Where(condition, params...).
		Limit(r.limit).
		Offset(r.skip).Count(&count)

	if cntResult.Error != nil {
		return 0, result.Error
	}

	return int(count), result.Error
}

func (r *gormRepos) Count(model interface{}) (int, error) {
	var count int64
	result := r.db.Model(model).Count(&count)
	return int(count), result.Error
}

func (r *gormRepos) Exists(model interface{}, params ...interface{}) (bool, error) {
	if len(params) > 1 {
		result := r.db.Limit(1).Where(params[0], params[1]).Find(model)
		return result.RowsAffected > 0, result.Error
	} else {
		result := r.db.First(model, params)
		return result.RowsAffected > 0, result.Error
	}

}

// Raw 执行SQL语句并返回结果集
func (r *gormRepos) Raw(model interface{}, sql string, params ...interface{}) error {
	result := r.db.Raw(sql, params...).Find(model)
	return result.Error
}

// Exec 执行原生的SQL语句，不会返回执行结果
func (r *gormRepos) Exec(sql string, params ...interface{}) error {
	result := r.db.Exec(sql, params...)
	return result.Error
}

func (r *gormRepos) Setup(models ...interface{}) error {
	if r.db.Name() == "postgres" {
		r.db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	}
	return r.db.AutoMigrate(models...)
}

func (r *gormRepos) Clear(models ...interface{}) error {
	for _, t := range models {
		re := r.db.Where("1=1").Delete(t)
		if re.Error != nil {
			return re.Error
		}
	}
	return nil
}

func (r *gormRepos) DB() interface{} {
	return r.db
}
