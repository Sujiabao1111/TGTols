
// 自动生成模板DailyUserStats
package example
import (
	"time"
)

// dailyUserStats表 结构体  DailyUserStats
type DailyUserStats struct {
  Id  *int `json:"id" form:"id" gorm:"primarykey;column:id;size:20;"`  //id字段
  UserId  *int `json:"userId" form:"userId" gorm:"comment:用户ID;column:user_id;size:20;"`  //用户ID
  StatDate  *time.Time `json:"statDate" form:"statDate" gorm:"comment:统计日期;column:stat_date;"`  //统计日期
  BetAmount  *float64 `json:"betAmount" form:"betAmount" gorm:"comment:当日总流水/打码量;column:bet_amount;size:15;"`  //当日总流水/打码量
  WinAmount  *float64 `json:"winAmount" form:"winAmount" gorm:"comment:当日总输赢(派彩-下注);column:win_amount;size:15;"`  //当日总输赢(派彩-下注)
  DepositAmount  *float64 `json:"depositAmount" form:"depositAmount" gorm:"comment:当日总充值;column:deposit_amount;size:15;"`  //当日总充值
  WithdrawAmount  *float64 `json:"withdrawAmount" form:"withdrawAmount" gorm:"comment:当日总提现;column:withdraw_amount;size:15;"`  //当日总提现
  LoginCount  *int `json:"loginCount" form:"loginCount" gorm:"comment:当日登录次数;column:login_count;size:10;"`  //当日登录次数
  CreatedAt  *time.Time `json:"createdAt" form:"createdAt" gorm:"column:created_at;"`  //createdAt字段
  UpdatedAt  *time.Time `json:"updatedAt" form:"updatedAt" gorm:"column:updated_at;"`  //updatedAt字段
}


// TableName dailyUserStats表 DailyUserStats自定义表名 daily_user_stats
func (DailyUserStats) TableName() string {
    return "daily_user_stats"
}





