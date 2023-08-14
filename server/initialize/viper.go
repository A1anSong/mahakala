package initialize

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"server/global"
)

func Viper() {
	v := viper.New()
	v.SetConfigFile("config.yaml")
	v.SetConfigType("yaml")
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("致命错误：配置文件读取失败: %s \n", err))
	}
	v.WatchConfig()

	v.OnConfigChange(func(e fsnotify.Event) {
		if err = v.Unmarshal(&global.Config); err != nil {
			fmt.Println(err)
		}
	})

	if err = v.Unmarshal(&global.Config); err != nil {
		fmt.Println(err)
	}

	global.Viper = v
	err = global.Viper.WriteConfig()
	if err != nil {
		fmt.Println(err)
	}
}
