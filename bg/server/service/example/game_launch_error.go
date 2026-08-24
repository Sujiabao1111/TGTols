package example

import (
	"context"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	model "github.com/flipped-aurora/gin-vue-admin/server/model/example"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
)

type GameLaunchErrorService struct{}

func (gameLaunchErrorService *GameLaunchErrorService) GetGameLaunchErrorList(ctx context.Context, info exampleReq.GameLaunchErrorSearch) (list []model.GameLaunchErrorRecord, total int64, err error) {
	if info.UserID == nil || *info.UserID == 0 {
		return []model.GameLaunchErrorRecord{}, 0, nil
	}

	db := global.GVA_DB.WithContext(ctx).Table("user_game_launch_errors e").
		Select(`
			e.id,
			e.user_id,
			COALESCE(NULLIF(e.username, ''), u.username) AS username,
			e.game_code,
			e.game_name,
			e.provider_code,
			e.provider_name,
			e.is_lobby,
			e.is_mobile,
			e.language,
			e.error_type,
			e.error_message,
			e.page_url,
			e.game_url_host,
			e.client_ip,
			e.user_agent,
			e.created_at
		`).
		Joins("LEFT JOIN users u ON u.id = e.user_id").
		Where("e.user_id = ?", *info.UserID)

	if errorType := strings.TrimSpace(info.ErrorType); errorType != "" {
		db = db.Where("e.error_type = ?", errorType)
	}

	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err = db.Scopes(info.Paginate()).
		Order("e.created_at DESC").
		Scan(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
