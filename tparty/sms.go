package tparty

import (
	// "fmt"
	"encoding/json"
	"errors"
	"github.com/valyala/fasthttp"
	"log"
	"math/rand"
	"strconv"
	"time"
	// "crypto/hmac"
	// "encoding/base64"
	// "crypto/sha1"
)

const (
	// SendTextURI 短信发送接口
	SendTextURI = "https://sms-api.upyun.com/api/messages"
	// TextTemplateID 短信模板编号
	TextTemplateID = "541"
)

// SendTextTo call upyun sms service
func SendTextTo(num string) (string, error) {
	var client fasthttp.Client
	var req fasthttp.Request
	var resp fasthttp.Response
	req.Header.SetMethod("POST")
	token := "4q07FNCDKKZKx5g83nXjIJwSVn4yIl"
	req.Header.Set("Authorization", token)
	req.Header.SetRequestURI(SendTextURI)
	req.Header.SetContentType("application/json")
	code := genConfirmCode()
	var rowData = map[string]string{"mobile": num, "template_id": TextTemplateID, "vars": code}
	jsonData, err := json.Marshal(rowData)
	if err != nil {
		log.Printf("Produce jsonData inner func:SendTextTo failed, for %s\n\r", err.Error())
		return "", err
	}
	req.SetBody(jsonData)
	err = client.Do(&req, &resp)
	if err != nil {
		log.Printf("Send text failed, for %s\n\r", err.Error())
		return "", err
	}
	if resp.StatusCode() != 200 {
		log.Printf("Send text failed, for %s\n\r", string(resp.Body()))
		return "", errors.New("Wrong request to UpYun SMS-API")
	}
	return code, nil
}

// 无需考虑线程安全问题
var count int

func genConfirmCode() string {
	count++
	// 因为种子会改变，所以才有线程安全问题
	rand.Seed(time.Now().UnixNano())
	add1 := rand.Intn(5) + 4
	add2 := rand.Intn(7876) + 1
	add3 := count % 1000
	plus := add1*10000 + add2 + add3
	// string 会解码数据
	return strconv.Itoa(plus)
}

// for test
// func genConfirmCode() string {
// 	count++
// 	// 因为种子会改变，所以才有线程安全问题
// 	rand.Seed(time.Now().UnixNano())
// 	add1 := rand.Intn(5) + 4
// 	add2 := rand.Intn(7876) + 1
// 	add3 := count % 1000
// 	plus := add1*10000 + add2 + add3
// 	// string 会解码数据
// 	strconv.Itoa(plus)
// 	return "33305"
// }
