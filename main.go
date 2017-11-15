// package main

// import (
// 	"github.com/buaazp/fasthttprouter"
// 	"github.com/valyala/fasthttp"
// 	"log"
// 	"gos/handler"
// )

// func registerURL() *fasthttprouter.Router{
// 	// URI更新
// 	r := fasthttprouter.New()
// 	// 注册
// 	r.POST("/users", handler.UserRegister)
// 	// 登录
// 	r.POST("/sessions", handler.UserLogin)
// 	r.GET("/sessions", handler.ResumeSession)
// 	// 发送验证码
// 	r.POST("/codes", handler.SendConfirmCode)
// 	// 核对验证码
// 	r.GET("/codes", handler.CheckRegisterCode)

// 	return r
// }

// func runServer() error {
// 	r := registerURL()
// 	err := fasthttp.ListenAndServe(":8080", r.Handler)
// 	return err
// }

// func main() {
// 	log.Println("Server started...")
// 	err := runServer()
// 	if err != nil {
// 		log.Fatalln("Server stoped, for " + err.Error())
// 	}
// }

// // todo
// // 1. global logger for each request
// // 2. persistence for confirm-code and user-session

// // ----------------------

// // type People struct {
// // 	name string
// // }

// // func temp() *People{
// // 	var bob People
// // 	bob.name = "Bob"
// // 	return &bob
// // }

// // func variable(num... int)  {
// // 	log.Printf("%d\n", len(num))
// // }

package main

import (
	// "log"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"gos/dbc"
	"fmt"
	"strconv"
	"html/template"
)

func Counter(ctx *fasthttp.RequestCtx) {
	if  ctx.Request.Header.IsGet() {
	v, _ := dbc.GetKeyValue("counter")
	n, _ := strconv.Atoi(v)
	dbc.SetKeyValue("counter", strconv.Itoa(n+1))
	fmt.Fprintf(ctx, "times: %d", n+1)
	}
}

func MakeUser(ctx *fasthttp.RequestCtx) {
	if ctx.Request.Header.IsPost() {
	phone := string(ctx.PostArgs().Peek("num"))
	password := string(ctx.PostArgs().Peek("password"))
	nickname := string(ctx.PostArgs().Peek("nickname"))
	dbc.AppendUser(phone, password, nickname)
	ctx.Response.SetStatusCode(301)
	ctx.Response.Header.Set("Location", "/user/"+phone)
	}
}

func GetUser(ctx *fasthttp.RequestCtx) {
	if ctx.Request.Header.IsGet() {
	phone := ctx.UserValue("num").(string)
	info, _ := dbc.QueryUser(phone)
	fmt.Fprintf(ctx, "%s\n", info[0])
	fmt.Fprintf(ctx, "%s", info[1])
	}
}

func FormPage(ctx *fasthttp.RequestCtx) {
	t, _ := template.ParseFiles("./form.html")
	ctx.SetContentType("text/html")
	t.Execute(ctx, nil)
}

func main() {
	var r fasthttprouter.Router
	r.GET("/", FormPage)
	r.GET("/counter", Counter)
	r.POST("/user", MakeUser)
	r.GET("/user/:num", GetUser)

	fasthttp.ListenAndServe("0.0.0.0:8080", r.Handler)
}

