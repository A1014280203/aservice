package handler


import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"log"
	"time"
	"gos/dbc"
	"gos/tparty"
	"gos/model"
	"strings"
	// for temp blocks
)


const (
	formatErrorCode  = 406
	unauthorizedCode 	= 401
	notFoundCode        = 404
	methodNotAllowdCode = 405
	conflictErrorCode = 409
	internalErrorCode   = 500
	internalError = `{"error":"Service not available now, try later"}`
	successCode			= 200
	createSuccessCode = 201
	asyncSuccessCode = 202
	deleteSuccessCode 	= 204
)

// 这些应该使用redis数据库
var numToCode map[string]string

func init() {
	numToCode = make(map[string]string)
}

// UserRegister user register REST API NOT view function
// data format:
// {"num":phone number, "password":raw password, "nickname":[nickname, ""], "code":confirm code}
// response is 'bodyless'
func UserRegister(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType("application/json")
	if ctx.Request.Header.IsPost() {
		var data map[string]string
		if err := JSONDecode(ctx, &data);err != nil{
			return
		}
		if data["num"] == "" || data["password"] == "" || data["code"] != numToCode[data["num"]] {
			log.Printf("Receive bad register data %v\n", data)
			ctx.Response.Header.SetStatusCode(formatErrorCode)
			ctx.Response.AppendBody(httpException("Invalid Filed Found"))
			return
		}
		// check duplicate
		info, err := dbc.QueryUser(data["num"])
		if err != nil {
			ctx.Response.Header.SetStatusCode(internalErrorCode)
			ctx.Response.AppendBodyString(internalError)
			return
		}
		if info[0] != "" {
			ctx.Response.Header.SetStatusCode(conflictErrorCode)
			ctx.Response.AppendBody(httpException("User Already Existed"))
			return
		}
		// save to database
		err = dbc.AppendUser(data["num"], data["password"], data["nickname"])
		if err != nil {
			ctx.Response.Header.SetStatusCode(internalErrorCode)
			ctx.Response.AppendBodyString(internalError)
			return
		}
		ctx.Response.Header.SetStatusCode(createSuccessCode)
		ctx.Response.AppendBody(ctx.Request.Body())
		return
	}
	ctx.Response.Header.SetStatusCode(methodNotAllowdCode)
}

// SendConfirmCode will return internalError if the phone number is bad.
// {"num":phone-num}
func SendConfirmCode(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType("application/json")
	if ctx.Request.Header.IsPost() {
		var data map[string]string
		if err := JSONDecode(ctx, &data);err != nil {
			return
		}
		if data["num"] == "" {
			log.Printf("Receive empty phone number for sending code\n")
			ctx.Response.Header.SetStatusCode(formatErrorCode)
			ctx.Response.AppendBody(httpException("Filed 'num' should not be empty"))
			return
		}
		code, err := tparty.SendTextTo(data["num"])
		if err != nil {
			ctx.Response.Header.SetStatusCode(internalErrorCode)
			ctx.Response.AppendBodyString(internalError)
			return
		}
		numToCode[data["num"]] = code
		ctx.Response.Header.SetStatusCode(asyncSuccessCode)
		ctx.Response.AppendBody(ctx.Request.Body())
		return
	}
	ctx.Response.Header.SetStatusCode(methodNotAllowdCode)
}

// CheckRegisterCode requires 
// /codes/{num}
// {"code": confirm code}
func CheckRegisterCode(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType("application/json")
	if ctx.Request.Header.IsGet() {
		var data map[string]string
		num := ctx.UserValue("num").(string)
		if err := JSONDecode(ctx, &data); err != nil {
			return
		}
		if num == "" || data["code"] == "" {
			log.Printf("Receive bad code check data %v\n", data)
			ctx.Response.Header.SetStatusCode(formatErrorCode)
			ctx.Response.AppendBody(httpException("Empty Filed Found"))
			return
		}
		if numToCode[num] != data["code"] {
			ctx.Response.Header.SetStatusCode(unauthorizedCode)
			ctx.Response.AppendBody(httpException("Comfirm Failed"))
			return 
		}
		ctx.Response.Header.SetStatusCode(successCode)
		ctx.Response.AppendBody(ctx.Request.Body())
		return
	}
	ctx.Response.Header.SetStatusCode(methodNotAllowdCode)
}

// UserLogin data format:
// receive {"num":phone number, "password":raw password}
// return {"num":phone number, "nickname":user nickname}
func UserLogin(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType("application/json")
	if ctx.Request.Header.IsPost() {
		var data map[string]string
		if err := JSONDecode(ctx, &data);err != nil {
			return
		}
		if data["num"] == "" || data["password"] == "" {
			log.Printf("Receive empty login data %v\n", data)
			ctx.Response.Header.SetStatusCode(formatErrorCode)
			ctx.Response.AppendBody(httpException("Empty Filed Found"))
			return
		}
		// query from database
		info, err := dbc.QueryUser(data["num"])
		if err != nil {
			ctx.Response.Header.SetStatusCode(internalErrorCode)
			ctx.Response.AppendBodyString(internalError)
			return
		}
		if info[0] == "" {
			ctx.Response.Header.SetStatusCode(notFoundCode)
			ctx.Response.AppendBody(httpException("No Such User"))
			return
		}
		if !dbc.CheckHashPassword(data["num"], data["password"], info[0]) {
			ctx.Response.Header.SetStatusCode(unauthorizedCode)
			ctx.Response.AppendBody(httpException("Wrong Password"))
			return
		}
		// process login details
		// make response
		ctx.Response.Header.SetStatusCode(successCode)
		respData, _ := json.Marshal(map[string]string{"num": data["num"], "nickname": info[1]})
		ctx.Response.AppendBody(respData)
		// 旧版session保存状态的实现
		// set cookie and session related
		// s := model.Sessions.CreateSession()
		// s.SetKV("num", data["num"])
		// s.SetKV("nickname", info[1])
		// c := makeCookie("sid", s.GetSessionID(), 3600*24*3)
		// ctx.Response.Header.SetCookie(c)
		// log.Printf("Create session[%s] for user[%s]\n", s.GetSessionID(), data["num"])

		// 现在使用JWT
		jwt := model.JWTManager.MakeJWT(map[string]string{}, map[string]string{"nickname":info[1]})
		ctx.Response.Header.Set("Token", jwt)
		log.Printf("Create JWT[%s] for user[%s]\n", jwt, data["num"])
		return
	}
	ctx.Response.Header.SetStatusCode(methodNotAllowdCode)
}

// ResumeSession allows user to login with cookie key 'sid'
// URI: /session/{num} GET
// Note: the var {num} is just to let the logger know who called this interface. It won't be checked.
func ResumeSession(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType("application/json")
	if ctx.Request.Header.IsGet() {
		// 旧版session保存状态的实现
		// sid := string(ctx.Request.Header.Cookie("sid"))
		// s := model.Sessions.GetSessionBySid(sid)
		// if sid == "" || s == nil {
		// 	ctx.Response.Header.SetStatusCode(notFoundCode)
		// 	ctx.Response.AppendBody(httpException("No Related Session Here"))
		// 	return
		// }
		// ctx.Response.Header.SetStatusCode(successCode)
		// ctx.Response.AppendBody(s.ToJSON())
		// log.Printf("Reuse session[%s] for user[%s]\n", s.GetSessionID(), s.GetValue("num"))
		// 现在使用JWT
		
		log.Printf("header: %s\n", ctx.Request.Header.Header())
		auth := string(ctx.Request.Header.Peek("Authorization"))
		log.Printf("auth: %s\n", auth)
		if auth == "" {
			ctx.Response.Header.SetStatusCode(unauthorizedCode)
			ctx.Response.AppendBody(httpException("Authorization cannot be empty"))
			return
		}
		parts := strings.Split(auth, " ")
		if len(parts) != 2 {
			ctx.Response.Header.SetStatusCode(formatErrorCode)
			ctx.Response.AppendBody(httpException("Authorization content format error"))
			return
		}
		dsrp := parts[0]
		jwt := parts[1]
		if dsrp != "Bearer" {
			ctx.Response.Header.SetStatusCode(formatErrorCode)
			ctx.Response.AppendBody(httpException("Authorization begins with 'Bearer '."))
			return
		}
		if !model.JWTManager.ValidJWT(jwt) {
			ctx.Response.Header.SetStatusCode(unauthorizedCode)
			ctx.Response.AppendBody(httpException("Invalid JSON Web Token"))
			return
		}
		log.Printf("Reuse JWT[%s]\n", jwt)
		return
	}
	ctx.Response.Header.SetStatusCode(methodNotAllowdCode)
}

// temp blocks
func httpException(stmt string) []byte {
	data := map[string]string{"error": stmt}
	b, _ := json.Marshal(data)
	return b
}

func JSONDecode(ctx *fasthttp.RequestCtx, data *map[string]string) error {
	err := json.Unmarshal(ctx.Request.Body(), data)
	if err != nil {
		log.Printf("Decode request body with JSON failed, for %s\n", err.Error())
		ctx.Response.Header.SetStatusCode(formatErrorCode)
		ctx.Response.AppendBody(httpException("Invalid Format, Need JSON"))
	}
	return err
}

func makeCookie(k, v string, secFromNow int64) *fasthttp.Cookie {
	var c fasthttp.Cookie
	nano := time.Duration(secFromNow*1000*1000*1000)
	exp := time.Now().Add(nano)
	c.SetKey(k)
	c.SetValue(v)
	// c.SetSecure(true) for https
	c.SetExpire(exp)
	return &c
}
