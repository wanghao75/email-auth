package conf

import (
	"fmt"
	"github.com/spf13/viper"
)

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/go/src/email-auth/conf")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("err is %v", err)
		panic("init config err")
	}
}
