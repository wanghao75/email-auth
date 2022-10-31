package controllers

import (
	"github.com/gin-gonic/gin"
)

type Gin struct {
	C *gin.Context
}

type Response struct {
	StatusCode int         `json:"status_code"`
	Msg        string      `json:"msg"`
	Data       interface{} `json:"data"`
}

func (g *Gin) Response(code int, msg string, data interface{}) {
	g.C.JSON(code, Response{
		StatusCode: code,
		Msg:        msg,
		Data:       data,
	})
}
