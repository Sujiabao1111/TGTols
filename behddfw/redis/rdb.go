package redis

import (
	"context"
	"gogogo/helpers"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/valyala/fastjson"
)

// redis DB .
type RDBHolder struct {
	C   redis.UniversalClient
	Ctx context.Context
}

var instanceRdb *RDBHolder
var onceRdb sync.Once

func GetRDbInstance() *RDBHolder {
	onceRdb.Do(func() {
		instanceRdb = new(RDBHolder)
	})
	return instanceRdb
}

// NewDB 初始化数据库
func (dbw *RDBHolder) NewRDB(url string, db int) {
	dbw.Ctx = context.Background()
	dbw.C = redis.NewClient(&redis.Options{
		Addr:     url,
		Password: "", // no password set
		DB:       db, // use default DB
	})
}

func (dbw *RDBHolder) StoreTicket(k string, v interface{}, exp time.Duration) error {
	return dbw.C.Set(dbw.Ctx, k, v, exp).Err()
}

func (dbw *RDBHolder) FindTicket(k string) (string, error) {
	return dbw.C.Get(dbw.Ctx, k).Result()
}

func (dbw *RDBHolder) DelTicket(k string) (err error) {
	if err := dbw.C.Del(dbw.Ctx, k).Err(); err != nil {
		return err
	}
	return nil
}

// 订阅/发布
// subscriber 订阅频道并接收消息
const (
	BgSubChannel = "sub_v"
	// 用户API区
	MSG1 = 1 // 下发配置更改
)

func subscriber(ctx context.Context, rdb *redis.UniversalClient) {
	pubsub := (*rdb).Subscribe(ctx, BgSubChannel)
	defer pubsub.Close()
	ch := pubsub.Channel()
	for msg := range ch {
		cmd := fastjson.GetInt([]byte(msg.Payload), "msgid")
		if cmd == 0 {
			continue
		}
		handler(cmd, msg.Payload)
	}
}

func handler(cmd int, payload string) bool {
	switch cmd {
	case MSG1:
		return handlemsg1(payload)
	default:
		return true
	}
}

func handlemsg1(payload string) bool {
	helpers.GetCfgInstance().LoadFromDb()
	return true
}

func InitPubSub() {
	ctx := context.Background()
	// 启动订阅者
	go subscriber(ctx, &GetRDbInstance().C)
}
