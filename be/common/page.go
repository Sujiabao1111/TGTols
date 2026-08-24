package common

import (
	"errors"
	"math"

	"gorm.io/gorm"
)

type PageParam struct {
	PageNo   int `json:"pageNo"`   // 页码
	PageSize int `json:"pageSize"` // 每页大小
}

func (page *PageParam) setDefault() {
	if page.PageNo <= 0 {
		page.PageNo = 1
	}
	switch {
	case page.PageSize > 100:
		page.PageSize = 100
	case page.PageSize <= 0:
		page.PageSize = 10
	}
}

type PageResp struct {
	PageNo    int `json:"page_no"`    // 页码
	PageSize  int `json:"page_size"`  // 每页大小
	Total     int `json:"total"`      // 數據總量
	PageCount int `json:"page_count"` // 總頁數
}

func FindWithPage(db *gorm.DB, page *PageParam, dest interface{}, useDeletedAt ...bool) (*PageResp, error) {
	if db == nil || page == nil || dest == nil {
		return nil, errors.New("FindWithPage param error")
	}
	page.setDefault()

	var (
		total     int64
		pageCount int
		err       error
	)
	if len(useDeletedAt) > 0 && useDeletedAt[0] {
		db = db.Unscoped()
	}
	if err = db.Count(&total).Error; err != nil {
		return nil, err
	}
	if page.PageSize != 0 {
		pageCount = int(math.Ceil(float64(total) / float64(page.PageSize)))
	}

	offset := (page.PageNo - 1) * page.PageSize
	db = db.Offset(offset).Limit(page.PageSize)

	if err = db.Find(dest).Error; err != nil {
		return nil, err
	}
	return &PageResp{
		PageNo:    page.PageNo,
		PageSize:  page.PageSize,
		Total:     int(total),
		PageCount: pageCount,
	}, nil
}

func FindAll(db *gorm.DB, dest interface{}, conds ...interface{}) error {
	if db == nil || dest == nil {
		return errors.New("FindAll param error")
	}

	var (
		err error
	)

	if err = db.Find(dest, conds...).Error; err != nil {
		return err
	}
	return nil
}

type BaseError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}
