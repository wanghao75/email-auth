package main

import (
	"email-auth/conf"
	"email-auth/controllers"
	"email-auth/cronjob"
	"email-auth/models"
	"email-auth/routers"
	"fmt"
	"github.com/robfig/cron/v3"
	"time"
)

func main() {
	conf.InitConfig()
	models.InitDB()
	controllers.InitRandStringList(controllers.RandStringList)
	loc, _ := time.LoadLocation("Asia/Shanghai")
	c := cron.New(cron.WithLocation(loc))
	go cronjob.Job(c)
	r := routers.RouterEmailHelper()
	if err := r.Run(); err != nil {
		fmt.Println("startup service failed, err:%v\n", err)
	}
}
