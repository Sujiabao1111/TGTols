package models

import "gogogo/models/dtos"

type DbConfigObject interface {
	*dtos.Config | *dtos.Sku
}

func LoadSomeConfigs[T DbConfigObject]() ([]T, error) {
	var u1 []T
	result := GetInstance().DbInstance.Find(&u1)
	return u1, result.Error
}
