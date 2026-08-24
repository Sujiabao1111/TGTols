package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gogogo/controllers"
	"gogogo/helpers"
	"gogogo/models"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"

	jwtware "github.com/gofiber/contrib/jwt"
)

// get the exit signal and handle last
func gracefulExitWeb(server *fiber.App) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	sig := <-ch

	fmt.Println("got a signal", sig)
	now := time.Now()
	err := server.Shutdown()
	if err != nil {
		fmt.Println("err", err)
	}

	// 实际退出所耗费的时间
	fmt.Println("------exited--------", time.Since(now))
}

var c *cron.Cron

func Setup() {
	// 0.设置随机种子
	// rand.Seed(time.Now().UnixNano())
	// use r := rand.New(rand.NewSource(time.Now().UnixNano())) to get rand
	// 1.12默认有这个不用写了
	runtime.GOMAXPROCS(runtime.NumCPU())
	// 1.读取配置文件(本地)
	// 1.x 从配置中心读取配置
	// 1.1.x 根据bin目录下的env读取相应的配置
	err := godotenv.Load("./.env")
	dev := "config"
	if err == nil {
		isDev := os.Getenv("dev")
		if isDev == "1" {
			dev = "config.dev"
		}
	}
	helpers.GetCfgInstance().Load(dev, ".")
	// 2.设置日志
	// 2-1.设置日志中心
	helpers.GetLoggerInstance().InitLogger("./all.log", "info")

	// 3.init database Instance
	models.GetInstance().Connect(helpers.GetCfgInstance().Conf.DbDsn, helpers.GetCfgInstance().Conf.PoolIdle, helpers.GetCfgInstance().Conf.PoolMax)
	// 3.1 从数据库预加载一些配置
	helpers.GetCfgInstance().LoadFromDb()
	// 3.2 挂载redis,不需要注释它
	// redis.GetRDbInstance().NewRDB(helpers.GetCfgInstance().Conf.RedisNode, helpers.GetCfgInstance().Conf.RedisDb)
	// 3.2.2 redis订阅发布
	// redis.InitPubSub()
	// 4.初始化定时任务
	c = controllers.SetupAndGo()

	// 5.初始化http服务器
	app := fiber.New(fiber.Config{
		JSONEncoder: func(v interface{}) ([]byte, error) {
			var buf bytes.Buffer
			encoder := json.NewEncoder(&buf)

			// 核心设置：禁止 HTML 转义
			// 这样 & 就会显示为 &，而不是 \u0026
			encoder.SetEscapeHTML(false)

			err := encoder.Encode(v)
			// Encode 默认会在末尾加换行符，Fiber 通常不需要，去掉它
			return bytes.TrimRight(buf.Bytes(), "\n"), err
		},
	})
	// 错误处理
	app.Use(recover.New())
	// 跨域配置
	app.Use(cors.New())
	// 不用jwt的路由
	controllers.InitRouteTables(app)
	// JWT Middleware
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(helpers.GetCfgInstance().Conf.Jwt)},
	}))
	// 需要jwt的路由
	controllers.InitRouteTablesWithJWT(app)

	go gracefulExitWeb(app)

	log.Fatal(app.Listen(fmt.Sprintf(":%d", helpers.GetCfgInstance().Conf.Port)))

	controllers.StopAndClean(c)
}
