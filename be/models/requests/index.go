package requests

// LoginRequest 登录请求参数
type LoginRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	SiteDomain string `json:"site_domain"`
}

type RegisterRequest struct {
	SiteDomain string `json:"site_domain"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	InviteCode string `json:"invite_code"` // 推荐人的邀请码 (可选)
	ParentID   uint64 `json:"parent_id"`   // 推荐人的ID (可选, 优先级低于InviteCode)
}
