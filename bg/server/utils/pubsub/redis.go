package pubsub

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/redis/go-redis/v9"
	"github.com/valyala/fastjson"
)

const (
	BgSubChannel = "sub_v"
	BgPubChannel = "pub_v"

	MSG1 = 1 // 消息示例
)

type HubData struct {
	IdToStatus sync.Map // appid - last submit heartbeat 105 message time
}

var hub *HubData

type MsgBase struct {
	Msgid int `json:"msgid"`
}

type Msg1 struct {
	MsgBase
	Appid string `json:"appid"`
}

func handlemsg1(payload string) bool {
	var mb Msg1
	err := json.Unmarshal([]byte(payload), &mb)
	if err != nil {
		return false
	}
	return true
}

func handler(cmd int, payload string) bool {
	switch cmd {
	case MSG1:
		return handlemsg1(payload)
	default:
		return true
	}
}

// subscriber 订阅频道并接收消息
func subscriber(ctx context.Context, rdb *redis.UniversalClient) {
	pubsub := (*rdb).Subscribe(ctx, BgSubChannel)
	defer pubsub.Close()

	ch := pubsub.Channel()

	global.GVA_LOG.Info("Subscriber is waiting for messages...")
	for msg := range ch {
		// global.GVA_LOG.Info("Received message from channel %s: %s\n", zap.String("channel", msg.Channel), zap.String("payload", msg.Payload))
		cmd := fastjson.GetInt([]byte(msg.Payload), "msgid")
		if cmd == 0 {
			continue
		}
		handler(cmd, msg.Payload)
	}
}

// publisher 发布消息到频道
func Publisher(ctx context.Context, rdb *redis.UniversalClient, jsonstr string) {
	err := (*rdb).Publish(ctx, BgPubChannel, jsonstr).Err()
	if err != nil {
		global.GVA_LOG.Error("Failed to publish message: ")
	}
	// fmt.Println("All messages published.")
}

func InitPubSub(r *redis.UniversalClient) {
	ctx := context.Background()
	// rdb := redis.NewClient(&redis.Options{
	// 	Addr: "localhost:6379", // Redis 服务器地址
	// })
	hub = &HubData{
		IdToStatus: sync.Map{},
	}
	// 启动订阅者
	go subscriber(ctx, r)
	// 启动定时任务
	// var option []cron.Option
	// option = append(option, cron.WithSeconds())
	// // 清理DB定时任务
	// global.GVA_Timer.AddTaskByFunc("Do every 30secs", "*/30 * * * * *", func() {

	// }, "定时检测在", option...)

	// 启动发布者
	// go publisher(ctx, rdb)
}
