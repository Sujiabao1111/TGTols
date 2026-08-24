
package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	
)

type DailyUserStatsSearch struct{
      UserId  *int `json:"userId" form:"userId"` 
      StatDateRange  []string  `json:"statDateRange" form:"statDateRange[]"`
    request.PageInfo
}
