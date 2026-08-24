package models

import (
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// gentool -dsn "root:123456@tcp(localhost:3306)/star?charset=utf8mb4&parseTime=True&loc=Local" -tables "muggles,mugglecashouts,robots,stages,configs" -modelPkgName "dtos" -onlyModel -outPath "./models/dtos"

type DbWrapper struct {
	DbInstance *gorm.DB
}

func NewDb() *DbWrapper {
	obj := new(DbWrapper)
	return obj
}

var instance *DbWrapper
var once sync.Once

func GetInstance() *DbWrapper {
	once.Do(func() {
		instance = NewDb()
	})
	return instance
}

func init() {

}

func (dbw *DbWrapper) Connect(dsn string, poolIdle, poolMax int) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(err)
	}
	_innerdb, err2 := db.DB()
	if err2 != nil {
		panic(err2)
	}
	_innerdb.SetMaxIdleConns(poolIdle)
	_innerdb.SetMaxOpenConns(poolMax)
	_innerdb.SetConnMaxLifetime(time.Hour)

	dbw.DbInstance = db
	return db
}

func (db *DbWrapper) DoOne(username string) *User {
	u1 := new(User)
	db.DbInstance.Model(u1).First(&u1)
	return u1
}
