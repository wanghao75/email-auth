package cronjob

import (
	"email-auth/controllers"
	"email-auth/models"
	"github.com/robfig/cron/v3"
)

func Job(c *cron.Cron) {
	c.AddFunc("@every 5m", func() {
		var users []models.UserEmail
		models.DB.Find(&users)

		for _, u := range users {
			code := controllers.RandNumString()
			u.UpdateCode(code)
		}
	})

	//c.AddFunc("@every 20s", func() {
	//	var users []models.UserEmail
	//	models.DB.Find(&users)
	//
	//	for _, u := range users {
	//		token := emails.RandAllString()
	//		u.UpdateToken(token)
	//	}
	//})

	c.Start()
	select {}
}
