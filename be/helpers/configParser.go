package helpers

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

type ConfigHolder struct {
	Conf     *Config
	filename string
	filepath string

	SomeDbMem *DbCfgMem
}

var instance *ConfigHolder
var once sync.Once

func GetCfgInstance() *ConfigHolder {
	once.Do(func() {
		instance = new(ConfigHolder)
		instance.Conf = new(Config)
	})
	return instance
}

func (ch *ConfigHolder) Reload() {
	ch.Load(ch.filename, ch.filepath)
}

func (ch *ConfigHolder) Load(filenameWithoutExtension, filepath string) {
	viper.SetConfigName(filenameWithoutExtension) // 设置配置文件名 (不带后缀)
	viper.AddConfigPath(filepath)                 // 第一个搜索路径
	err := viper.ReadInConfig()                   // 读取配置数据
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s ", err))
	}
	ch.filename = filenameWithoutExtension
	ch.filepath = filepath
	viper.Unmarshal(ch.Conf) // 将配置信息绑定到结构体上
	// 监控配置文件变化
	// viper.WatchConfig()
	// viper.OnConfigChange(func(e fsnotify.Event) {
	// 	fmt.Println("配置发生变更：", e.Name)
	// 	ch.Reload()
	// })
}

func (ch *ConfigHolder) LoadFromDb() {
	ch.SomeDbMem = &DbCfgMem{}
}
