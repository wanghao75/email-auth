package controllers

import (
	"email-auth/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/util/sets"
	"math/rand"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	typeErrorCode = 601
	sendEmailErr  = 602
	createUserErr = 603
	sendSuccess   = 201
)

var CHARS = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
	"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}

func RandAllString(strLen int) string {
	str := strings.Builder{}
	length := len(CHARS)
	for i := 0; i < strLen; i++ {
		l := CHARS[rand.Intn(length)]
		str.WriteString(l)
	}
	return str.String()
}

func RandNumString() string {
	str := strings.Builder{}
	length := 10
	for i := 0; i < 6; i++ {
		str.WriteString(CHARS[52+rand.Intn(length)])
	}
	return str.String()
}

type user struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Code         string `json:"code"`
	State        string `json:"state"`
	ResponseType string `json:"response_type"`
	RedirectUri  string `json:"redirect_uri"`
}

var RandStringList = sets.NewString()

func InitRandStringList(l sets.String) {
	var users []models.UserEmail
	if err := models.DB.Find(&users).Error; err != nil {
		models.DB.Find(&users)
	}

	if len(users) != 0 {
		for _, u := range users {
			l.Insert(u.Token)
			l.Insert(u.RefreshToken)
			l.Insert(u.AuthCode)
		}
	}
}

func SendCode(c *gin.Context) {

	G := Gin{
		C: c,
	}
	var u user
	if err := G.C.ShouldBind(&u); err != nil {
		G.Response(http.StatusBadRequest, fmt.Sprintf("ShouldBind request body failed %v", err), "")
		return
	}

	if !VerifyEmailFormat(u.Email) {
		G.Response(http.StatusBadRequest, "invalid email address", "")
		return
	}

	code := RandNumString()
	token := ""
	refreshToken := ""
	state := u.State
	auth := ""

	for {
		a := RandAllString(10)
		t := RandAllString(20)
		r := RandAllString(20)
		if !RandStringList.Has(a) && !RandStringList.Has(t) && !RandStringList.Has(r) {
			auth = a
			token = t
			refreshToken = r
			break
		}
	}

	RandStringList.Insert(token)
	RandStringList.Insert(refreshToken)
	RandStringList.Insert(auth)

	userInfo := models.UserEmail{}
	if err := models.DB.Where("email = ?", u.Email).First(&userInfo).Error; err != nil {
		userInfo = models.UserEmail{UserName: u.Name, Email: u.Email, EmailCode: code,
			Token: token, RefreshToken: refreshToken,
			State: state, AuthCode: auth, TokenExpiry: 14400}

		if err3 := sendEmailCode(code, u.Email); err3 != nil {
			G.Response(sendEmailErr, fmt.Sprintf("email send code failed %v", err3), map[string]interface{}{})
			return
		}

		if err2 := userInfo.CreateUser(); err2 != nil {
			G.Response(createUserErr, fmt.Sprintf("create user failed %v", err2), map[string]interface{}{})
			return
		}
	} else {
		if err4 := sendEmailCode(userInfo.EmailCode, u.Email); err4 != nil {
			G.Response(sendEmailErr, fmt.Sprintf("email send code failed %v", err4), map[string]interface{}{})
			return
		}
	}

	G.Response(sendSuccess, "verify code has been sent", map[string]interface{}{})
	return
}

func ReSendCode(c *gin.Context) {
	G := Gin{
		C: c,
	}
	e := c.PostForm("email")
	u := &models.UserEmail{Email: e}

	emailCode := RandNumString()
	u.UpdateCode(emailCode)

	err := sendEmailCode(emailCode, e)
	if err != nil {
		G.Response(http.StatusBadRequest, fmt.Sprintf("email send code failed %v", err), map[string]interface{}{})
		return
	}

	G.Response(http.StatusOK, "resend code success", map[string]interface{}{})

	return
}

func sendEmailCode(code, emailAddress string) error {
	//host := viper.GetString("email.host")
	//if host == "" {
	//	host = os.Getenv("EMAIL_HOST")
	//}
	host := os.Getenv("EMAIL_HOST")

	//port := viper.GetString("email.port")
	//if port == "" {
	//	port = os.Getenv("EMAIL_PORT")
	//}
	port := os.Getenv("EMAIL_PORT")

	//us := viper.GetString("email.user")
	//if us == "" {
	//	us = os.Getenv("EMAIL_HOST_USER")
	//}
	us := os.Getenv("EMAIL_HOST_USER")

	//secret := viper.GetString("email.secret")
	//if secret == "" {
	//	secret = os.Getenv("EMAIL_HOST_PASSWORD")
	//}
	secret := os.Getenv("EMAIL_HOST_PASSWORD")

	auth := smtp.PlainAuth("", us, secret, host)
	contentType := "Content-Type: text/html; charset=UTF-8"
	subject := "登录验证码"
	body := fmt.Sprintf(`
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset=utf-8" />
			<title>MMOGA POWER</title>
		</head>
		<body>
			您的邮件验证码： %s, 5分钟内有效
		</body>
		</html>`, code)

	sendUserName := "Helios email-verify bot"

	msg := []byte("To: " + emailAddress + "\r\nFrom: " + sendUserName + "<" + us + ">" + "\r\nSubject: " + subject + "\r\n" + contentType + "\r\n\r\n" + body)

	err := smtp.SendMail(fmt.Sprintf("%s:%s", host, port), auth, us, []string{emailAddress}, msg)

	return err
}

func VerifyEmailCode(c *gin.Context) {
	G := Gin{C: c}
	var u user
	if err := G.C.ShouldBind(&u); err != nil {
		G.Response(http.StatusBadRequest, fmt.Sprintf("ShouldBind request body failed %v", err), map[string]interface{}{})
		return
	}

	var us models.UserEmail
	check := us.VerifyCode(u.Email, u.Code)
	if check {
		G.Response(http.StatusOK, "match record", map[string]interface{}{
			"email": us.Email,
		})
		return
	}

	G.Response(http.StatusNotFound, fmt.Sprintf("not find"), map[string]interface{}{
		"email": "",
	})
	return
}

func GetAuthCode(c *gin.Context) {
	G := Gin{C: c}
	responseType := G.C.Query("response_type")
	redirectUri := G.C.Query("redirect_uri")
	state := G.C.Query("state")
	email := G.C.Query("email")

	//redirectUri, _ = url.QueryUnescape(redirectUri)
	//email, _ = url.QueryUnescape(email)

	if responseType != "code" {
		G.Response(typeErrorCode, fmt.Sprintf("response_type: %s not illegal", responseType), map[string]interface{}{})
		return
	}

	var us models.UserEmail
	if err := models.DB.Where("email = ?", email).First(&us).Error; err != nil {
		G.Response(http.StatusNotFound, "not find", map[string]interface{}{})
		return
	}

	if state == us.State {
		G.Response(http.StatusOK, "verify success", map[string]interface{}{
			"redirect_uri": redirectUri,
			"state":        state,
			"code":         us.AuthCode,
		})
		return
	}

	G.Response(http.StatusForbidden, "state not match", map[string]interface{}{})
	return
}

func GetResource(c *gin.Context) {
	G := Gin{C: c}
	t := c.Query("token")
	fmt.Println("token is ", t)

	var us models.UserEmail
	if err := models.DB.Where("token = ?", t).First(&us).Error; err != nil {
		G.Response(http.StatusNotFound, "can not find record", map[string]interface{}{})
		return
	}

	n := time.Now().Unix()
	expiry := us.TokenGetTime + us.TokenExpiry
	fmt.Println("++++++++++++++++++++++++++++", t, n, expiry, us.TokenGetTime, us.TokenExpiry)
	if n > expiry {
		G.Response(http.StatusForbidden, "token not match", map[string]interface{}{})
		return
	}

	if us.Email != "" && us.UserName != "" {
		G.Response(http.StatusOK, "get user information success", map[string]interface{}{
			"email": us.Email,
			"name":  us.UserName,
		})
		return
	}
}

// RefreshTokenByRF
// PUT
// refresh token by refresh_token
func RefreshTokenByRF(c *gin.Context) {
	G := Gin{C: c}
	G.C.Header("Cache-Control", "no-store")
	t := c.Query("refresh_token")
	e := c.Query("email")
	grantType := c.Query("grant_type")

	if grantType != "refresh_token" {
		G.Response(typeErrorCode, fmt.Sprintf("grant_type: %s not illegal", grantType), map[string]interface{}{})
		return
	}

	us := models.UserEmail{RefreshToken: t, Email: e}

	newToken := ""
	for {
		newToken = RandAllString(20)
		if RandStringList.Has(newToken) {
			continue
		} else {
			break
		}
	}
	us.UpdateToken(newToken)

	G.Response(http.StatusOK, "", map[string]interface{}{
		"token":         us.Token,
		"refresh_token": t,
	})
	return
}

func GetTokenByCode(c *gin.Context) {
	G := Gin{C: c}
	grantType := c.Query("grant_type")
	redirectUri := c.Query("redirect_uri")
	code := c.Query("code")
	G.C.Header("Cache-Control", "no-store")

	if grantType != "authorization_code" {
		G.Response(typeErrorCode, fmt.Sprintf("grant_type: %s not illegal", grantType), map[string]interface{}{})
		return
	}

	var u models.UserEmail
	if err := models.DB.Where("auth_code = ?", code).First(&u).Error; err != nil {
		G.Response(http.StatusNotFound, "get token failed", map[string]interface{}{})
		return
	}

	t := time.Now().Unix()
	err := models.DB.Model(&models.UserEmail{}).Where("email = ?", u.Email).Update("token_get_time", t).Error
	fmt.Println(err)

	G.Response(http.StatusOK, "get token success", map[string]interface{}{
		"redirect_uri":  redirectUri,
		"token":         u.Token,
		"refresh_token": u.RefreshToken,
		// "expires_in":    36000,
	})

	return
}

func VerifyEmailFormat(email string) bool {
	pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`
	reg := regexp.MustCompile(pattern)
	fmt.Println("email match == ", reg.MatchString(email))
	return reg.MatchString(email)
}
