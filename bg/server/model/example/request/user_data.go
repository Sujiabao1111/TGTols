package request

import "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type UserDataSearch struct {
	UserID            *uint64  `json:"userId" form:"userId"`
	Username          string   `json:"username" form:"username"`
	RegisterIP        string   `json:"registerIp" form:"registerIp"`
	RegisterDomain    string   `json:"registerDomain" form:"registerDomain"`
	RegisterDate      string   `json:"registerDate" form:"registerDate"`
	RegisterDateRange []string `json:"registerDateRange" form:"registerDateRange[]"`
	RechargeDate      string   `json:"rechargeDate" form:"rechargeDate"`
	RechargeDateRange []string `json:"rechargeDateRange" form:"rechargeDateRange[]"`
	request.PageInfo
}
