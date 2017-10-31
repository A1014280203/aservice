package main

import (
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"log"
	"gos/handler"
)

func registerURL() *fasthttprouter.Router{
	r := fasthttprouter.New()
	// 注册
	r.POST("/users", handler.UserRegister)
	// 登录
	r.POST("/sessions", handler.UserLogin)
	r.GET("/sessions/:num", handler.ResumeSession)
	// 发送验证码
	r.POST("/codes", handler.SendConfirmCode)
	// 核对验证码
	r.GET("/codes/:num", handler.CheckRegisterCode)

	return r
}

func runServer() error {
	r := registerURL()
	err := fasthttp.ListenAndServe(":8080", r.Handler)
	return err
}

func main() {
	log.Println("Server started...")
	err := runServer()
	if err != nil {
		log.Fatalln("Server stoped, for " + err.Error())
	}
}

// todo
// 1. global logger for each request
// 2. persistence for confirm-code and user-session