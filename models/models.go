package models

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"strings"
	"time"
)

type UserEmail struct {
	CreatedAt    time.Time
	UpdatedAt    time.Time
	UserName     string `gorm:"type:varchar(30)"`
	Email        string `gorm:"type:varchar(30);primary_key"`
	EmailCode    string `gorm:"type:varchar(6)"`
	AuthCode     string `gorm:"type:varchar(20);index"`
	Token        string `gorm:"type:varchar(20);index"`
	RefreshToken string `gorm:"type:varchar(20)"`
	State        string `gorm:"type:varchar(20)"`
	TokenExpiry  int64  `gorm:"type:int"`
	// TokenGetTime time.Time
	TokenGetTime int64 `gorm:"type:int"`
}

var DB *gorm.DB
var err error

func InitDB() {
	//dataBase := viper.GetString("datasource.database")
	//if dataBase == "" {
	//	dataBase = os.Getenv("DB_NAME")
	//}
	dataBase := os.Getenv("DB_NAME")

	//host := viper.GetString("datasource.host")
	//if host == "" {
	//	host = os.Getenv("DB_HOST")
	//}
	host := os.Getenv("DB_HOST")

	//port := viper.GetString("datasource.port")
	//if port == "" {
	//	port = os.Getenv("DB_PORT")
	//}
	port := os.Getenv("DB_PORT")

	//
	//userName := viper.GetString("datasource.username")
	//if userName == "" {
	//	userName = os.Getenv("DB_USER")
	//}
	userName := os.Getenv("DB_USER")

	//password := viper.GetString("datasource.password")
	//if password == "" {
	//	password = os.Getenv("DB_PASSWORD")
	//}
	password := os.Getenv("DB_PASSWORD")

	//charset := viper.GetString("datasource.charset")
	//if charset == "" {
	//	charset = os.Getenv("DB_CHARSET")
	//}
	charset := os.Getenv("DB_CHARSET")

	//loc := viper.GetString("datasource.loc")
	//if loc == "" {
	//	loc = os.Getenv("DB_LOC")
	//}
	loc := os.Getenv("DB_LOC")

	addr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=%s",
		userName, password, host, port, dataBase, charset, strings.Replace(loc, "/", "%2F", 1))
	DB, err = gorm.Open(mysql.Open(addr), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
		panic("init DB failed")
	}

	db, _ := DB.DB()
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)

	DB.AutoMigrate(&UserEmail{})
}

func (user *UserEmail) CreateUser() error {
	return DB.Create(&user).Error
}

func (user *UserEmail) UpdateCode(code string) {
	DB.Where("email = ?", user.Email).Find(&user)
	user.EmailCode = code

	DB.Save(user)
}

func (user *UserEmail) UpdateToken(token string) {
	DB.Where("email = ? AND refresh_token = ?", user.Email, user.RefreshToken).Find(&user)
	user.Token = token

	DB.Save(user)
}

func (user *UserEmail) VerifyCode(email, code string) bool {
	DB.Where("email = ?", email).First(&user)
	if user.EmailCode == code {
		return true
	}

	return false
}
