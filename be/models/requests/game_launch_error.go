package requests

// GameLaunchErrorReportRequest is sent by the frontend when third-party game entry fails.
type GameLaunchErrorReportRequest struct {
	GameCode     string `json:"game_code"`
	GameName     string `json:"game_name"`
	ProviderCode string `json:"provider_code"`
	ProviderName string `json:"provider_name"`
	IsLobby      bool   `json:"is_lobby"`
	IsMobile     bool   `json:"is_mobile"`
	Language     string `json:"language"`
	ErrorType    string `json:"error_type"`
	ErrorMessage string `json:"error_message"`
	PageURL      string `json:"page_url"`
	GameURL      string `json:"game_url"`
}
