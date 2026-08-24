package models

// GORM 使用结构体名的 蛇形命名 作为表名。对于结构体 User，根据约定，其表名为 users
type User struct {
	Id   int
	Name string
}
