
# tpl-go

Go API 开发脚手架

> 1. `config.json` 配置文件,在helpers/configs内,一些基于内存的配置需要自己定义,测试环境可选`config.dev.json`
> 2. 表 `Users` 对应 `models/dtos/user.go`,使用gorm的gentool工具生成,参考dbhandler.go
> 3. 默认db first,以数据库建表后gentool生成的数据结构为准,部署时可以打开sync将数据结构同步到生产服务器

- 路由使用 [fiber](https://github.com/gofiber/fiber)
- ORM使用 [gorm](https://github.com/go-gorm/gorm)
- db first,db to code[gorm-gentool](https://gorm.io/gen/gen_tool.html)
- Redis使用 [go-redis](https://github.com/redis/go-redis)
- 日志使用 [zap](https://github.com/uber-go/zap)
- 配置使用 [viper](https://github.com/spf13/viper)
- 命令行使用 [cobra](https://github.com/spf13/cobra)
- http请求使用 [Resty](https://github.com/go-resty/resty/)
- 工具包使用 [kbutils](https://github.com/kael777/kbutils)
- unique id generator使用 [xid](https://github.com/rs/xid)
- 包含jwt认证,请求日志,跨域,中间件
- 包含文件上传(支持分片上传)
- 图片URL访问(缩略图,裁切,标注)
- 简单好用的 API Result 统一输出方式
- 已加入randx部分多种随机数生成算法,时间轮/权重随机,等常用模块代码可以另外加入

本地配置使用config.dev.json不上传git
在.env里配置dev=1来获取config.dev.json配置

go mod tidy
go run main.go