package helpers

type Config struct {
	LocalURL    string
	Port        int
	DbDsn       string
	PoolIdle    int
	PoolMax     int
	Jwt         string
	RedisNode   string
	RedisDb     int
	Agentid     string
	Agentapi    string
	Prefix      string
	Gamehallapi string
}

type DbCfgMem struct {
	DbCfg struct{}
	Raw   string
}
