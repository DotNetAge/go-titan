package repository

// Repository 基于模式定义了通用的数据访问接口，用以解决一般的CRUD问题
type Repository interface {
	Limit(limit int) Repository

	Skip(skip int) Repository
	// Add 增加新的数据对象
	Add(model interface{}) error
	// Update 更新数据对象
	Update(model interface{}) error
	// Delete 删除指定的数据对象
	Delete(model interface{}, params ...interface{}) error
	// Get 获取指定主键值相同对象
	// model 为返回的对象, params为实体的ID值或值的列表
	Get(model interface{}, params ...interface{}) error
	// All 获取全部实体对象
	All(model interface{}) (int, error)
	// Query 查找符合指定条件的实体对象
	Query(model interface{}, condition string, params ...interface{}) (int,error)
	// Count 统计表内有多少个对象
	Count(model interface{}) (int, error)
	// Exists 返回符合条件的实体对象是否存在
	Exists(model interface{}, params ...interface{}) (bool, error)
	// Raw 执行SQL语句并返回结果集
	Raw(model interface{}, sql string, params ...interface{}) error
	// Exec 执行原生的SQL语句，不会返回执行结果
	Exec(sql string, params ...interface{}) error
	// Setup 初始化数据表
	Setup(model ...interface{}) error
	// Clear 清空数据表
	Clear(model ...interface{}) error
	// DB 返回仓库使用的数据库实例
	DB() interface{}
}

//DataWorks 统一数据存储对象(UnitOfWorks)
type DataWorks interface {
	// Repos 返回一个通用的可访问存储
	Repos() Repository

	// Setup 创建数据库结构
	Setup() error

	// Init 用于初始化数据库
	Init() error

	// Reset 重置所有相关数据至初始状态
	Reset() error
}
