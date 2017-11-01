package dbc

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strings"
)

var db *sql.DB
var conn redis.Conn
var err error

func init() {
	db, err = sql.Open("mysql", "")
	checkErr(err)
	conn, err = redis.Dial("", "")
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func nullToEmpty(value sql.NullString) string {
	if value.Valid {
		return value.String
	}
	return ""
}

func genHashPassword(phone, password string) string {
	s := sha256.New()
	s.Write([]byte(password))
	s.Write([]byte(phone))
	pwhash := s.Sum(nil)
	return hex.EncodeToString(pwhash)
}

func CheckHashPassword(phone, password, stored string) bool {
	s := sha256.New()
	s.Write([]byte(password))
	s.Write([]byte(phone))
	pwhash := s.Sum(nil)
	if strings.Compare(hex.EncodeToString(pwhash), stored) == 0 {
		return true
	}
	return false
}

func AppendUser(phone, password, nickname string) error {
	// 这里就会判断表中有没有这个字段
	stmt, err := db.Prepare("INSERT users SET phone=?,password=?,nickname=?")
	checkErr(err)
	_, err = stmt.Exec(phone, genHashPassword(phone, password), nickname)
	if err != nil {
		// log 自带时间部分
		log.Printf("Cannot append user[%s %s], for %s\n\r", phone, password, nickname, err.Error())
		return err
	}
	log.Printf("Append user[%s]\n\r", phone)
	return nil
}

func SetNickName(phone, nickname string) error {
	stmt, err := db.Prepare("UPDATE users set nickname=? where phone=?")
	checkErr(err)
	_, err = stmt.Exec(nickname, phone)
	if err != nil {
		log.Printf("Cannot set user[%s]'s nickname[%s], for %s\n\r", phone, nickname, err.Error())
		return err
	}
	log.Printf("Set user[%s] nickname[%s]\n\r", phone, nickname)
	return nil
}

func UpdatePassword(phone, newPassword string) error {
	stmt, err := db.Prepare("UPDATE users set password=? where phone=?")
	checkErr(err)
	_, err = stmt.Exec(newPassword, phone)
	if err != nil {
		log.Printf("Cannot change user[%s]'s password to new password[%s], for %s\n\r", phone, newPassword, err.Error())
		return err
	}
	log.Printf("Change user[%s]'s password\n\r", newPassword)
	return nil
}

func QueryUser(phone string) ([2]string, error) {
	sqlStmt := "SELECT * FROM users where phone='" + phone + "';"
	rows, err := db.Query(sqlStmt)
	checkErr(err)
	for rows.Next() {
		var phone string
		var password string
		var nickname sql.NullString
		err = rows.Scan(&phone, &password, &nickname)
		if err != nil {
			log.Printf("Find user[%s] failed\n\r, for %s", phone, err.Error())
			return [2]string{}, err
		}
		return [2]string{password, nullToEmpty(nickname)}, nil
	}
	log.Printf("Cannot find user[%s]\n\r", phone)
	return [2]string{"", ""}, nil
}

func SetKeyValue(k, v string) error {
	_, err := conn.Do("set", k, v)
	if err != nil {
		log.Printf("Cannot set(k=%s, v=%s) on redis, for %s\n\r", k, v, err.Error())
	}
	return err
}

func GetKeyValue(k string) (string, error) {
	rep, err := conn.Do("get", k)
	if err != nil {
		log.Printf("Connot get value of k=%s, for %s\n\r", k, err.Error())
		return "", err
	}
	return rep.(string), err
}
