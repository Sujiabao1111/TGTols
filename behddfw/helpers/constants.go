package helpers

const Version = "1.0.1"

const (
	_ = iota
	ErrorParseError
	ErrorSQLError
	ErrorOpenidError
)

const (
	ErrorParseErrorStr  = "数据解析错误"
	ErrorSQLErrorStr    = "数据库错误"
	ErrorOpenidErrorStr = "错误的openid"
)
