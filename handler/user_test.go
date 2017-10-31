package handler

import (
	"testing"
	"github.com/valyala/fasthttp"
	"encoding/json"
)

/*
1. set phoneNum
2. set globalCode
3. set database proprioty in dbc/dbc.go init()
4. switch genConfirmCode() to test special one in tpart/sms.go
5. make sure that genConfirmCode() return the same value as globalCode set
*/

const (
	phoneNum = "15968801476"
	// phoneNum = "15064757329"
)

var globalCode = "33305"
var globalSid = "sid"

func testSendConfirmCode(t *testing.T) {
	var ctx fasthttp.RequestCtx
	// supposed to succeed
	ctx.Request.Header.SetMethod("POST")
	ctx.Request.Header.SetContentType("application/json")
	ctx.Request.AppendBodyString(`{"num":"`+phoneNum+`"}`)
	//
	SendConfirmCode(&ctx)
	//
	t.Logf("===Test SendConfirmCode Result===\n")
	if ctx.Response.Header.StatusCode() - 200 > 99{
		t.Logf("StatusCode is %d\n", ctx.Response.Header.StatusCode())
		t.Fatalf("Response Data is %s\n", ctx.Response.Body())
		// os.Exit(1)
	}
	var data map[string]string
	if err := json.Unmarshal(ctx.Response.Body(), &data); err != nil {
		t.Logf("StatusCode is %d\n", ctx.Response.Header.StatusCode())
		t.Fatalf("Response Data is Not JSON Format %s\n", ctx.Response.Body())
		// os.Exit(1)
	}
	t.Logf("-StatusCode %d:\n", ctx.Response.Header.StatusCode())
	t.Logf("-Body %s\n", ctx.Response.Body())
}

func testCheckRegisterCode(t *testing.T) {
	var ctx fasthttp.RequestCtx
	// var code int
	// supposed to succeed
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.Header.SetContentType("application/json")
	ctx.SetUserValue("num", phoneNum)
	ctx.Request.AppendBodyString(`{"code":"`+globalCode+`"}`)
	// 
	CheckRegisterCode(&ctx)
	//
	t.Logf("===Test CheckRegisterCode Result===\n")
	if ctx.Response.Header.StatusCode() - 200 > 99{
		t.Logf("StatusCode is %d\n", ctx.Response.Header.StatusCode())
		t.Fatalf("Response Data is %s\n", ctx.Response.Body())
		// os.Exit(1)
		
	}
	var data map[string]string
	if err := json.Unmarshal(ctx.Response.Body(), &data); err != nil {
		t.Logf("StatusCode is %d\n", ctx.Response.Header.StatusCode())
		t.Fatalf("Response Data is Not JSON Format %s\n", ctx.Response.Body())
		// os.Exit(1)
	}
	t.Logf("-StatusCode %d:\n", ctx.Response.Header.StatusCode())
	t.Logf("-Body %s\n", ctx.Response.Body())
}

func testUserRegister(t *testing.T) {
	var ctx fasthttp.RequestCtx
	// supposed to succeed
	ctx.Request.Header.SetMethod("POST")
	ctx.Request.Header.SetContentType("application/json")
	ctx.Request.AppendBodyString(
		`{"num":"`+ phoneNum +`",
			"password":"123456",
			"nickname":"raka",
			"code":"`+ globalCode +`"}`)
	//
	UserRegister(&ctx)
	//
	t.Logf("===Test UserRegister Result===\n")
	if ctx.Response.Header.StatusCode() - 200 > 99{
		t.Logf("StatusCode is %d\n", ctx.Response.Header.StatusCode())
		t.Fatalf("Response Data is %s\n", ctx.Response.Body())
		// os.Exit(1)
		
	}
	var data map[string]string
	if err := json.Unmarshal(ctx.Response.Body(), &data); err != nil {
		t.Logf("StatusCode is %d\n", ctx.Response.Header.StatusCode())
		t.Fatalf("Response Data is Not JSON Format %s\n", ctx.Response.Body())
		// os.Exit(1)
	}
	t.Logf("-StatusCode %d:\n", ctx.Response.Header.StatusCode())
	t.Logf("-Body %s\n", ctx.Response.Body())
}

func testUserLogin(t *testing.T) {
	var ctx fasthttp.RequestCtx
	// supposed to succeed
	ctx.Request.Header.SetMethod("POST")
	ctx.Request.Header.SetContentType("application/json")
	ctx.Request.AppendBodyString(
		`{"num":"`+ phoneNum +`",
			"password":"123456"}`)
	//
	UserLogin(&ctx)
	//
	t.Logf("===Test UserLogin Result===\n")
	if ctx.Response.Header.StatusCode() - 200 > 99{
		t.Logf("StatusCode is %d\n", ctx.Response.Header.StatusCode())
		t.Fatalf("Response Data is %s\n", ctx.Response.Body())
		// os.Exit(1)
		
	}
	var data map[string]string
	if err := json.Unmarshal(ctx.Response.Body(), &data); err != nil {
		t.Logf("StatusCode is %d\n", ctx.Response.Header.StatusCode())
		t.Fatalf("Response Data is Not JSON Format %s\n", ctx.Response.Body())
		// os.Exit(1)
	}
	var c fasthttp.Cookie
	c.SetKey("sid")
	if !ctx.Response.Header.Cookie(&c) {
		t.Logf("Did not find cookie[sid] after login\n")
	}
	globalSid = string(c.Value())
	t.Logf("-StatusCode %d:\n", ctx.Response.Header.StatusCode())
	t.Logf("-Body %s\n", ctx.Response.Body())
}

func testResumeSession(t *testing.T) {
	var ctx fasthttp.RequestCtx
	// supposed to succeed
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.Header.SetContentType("application/json")
	ctx.SetUserValue("num", phoneNum)
	ctx.Request.Header.SetCookie("sid", globalSid)
	//
	ResumeSession(&ctx)
	//
	t.Logf("===Test ResumeSession Result===\n")
	if ctx.Response.Header.StatusCode() - 200 > 99{
		t.Logf("StatusCode is %d\n", ctx.Response.Header.StatusCode())
		t.Fatalf("Response Data is %s\n", ctx.Response.Body())
		// os.Exit(1)
		
	}
	var data map[string]string
	if err := json.Unmarshal(ctx.Response.Body(), &data); err != nil {
		t.Logf("StatusCode is %d\n", ctx.Response.Header.StatusCode())
		t.Fatalf("Response Data is Not JSON Format %s\n", ctx.Response.Body())
		// os.Exit(1)
	}
	t.Logf("-StatusCode %d:\n", ctx.Response.Header.StatusCode())
	t.Logf("-Body %s\n", ctx.Response.Body())
}


func Test_registe_login_reuseSession(t *testing.T) {
	testSendConfirmCode(t)
	testCheckRegisterCode(t)
	testUserRegister(t)
	testUserLogin(t)
	testResumeSession(t)
}
