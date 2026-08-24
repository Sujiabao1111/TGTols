
package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	
)

type StatsRetentionSearch struct{
      StatDateRange  []string  `json:"statDateRange" form:"statDateRange[]"`
    request.PageInfo
}
