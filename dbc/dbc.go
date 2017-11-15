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
		log.Printf("Cannot append user[%s %s], for %s\n", phone, password, nickname, err.Error())
		return err
	}
	log.Printf("Append user[%s]\n", phone)
	return nil
}

func SetNickName(phone, nickname string) error {
	stmt, err := db.Prepare("UPDATE users set nickname=? where phone=?")
	checkErr(err)
	_, err = stmt.Exec(nickname, phone)
	if err != nil {
		log.Printf("Cannot set user[%s]'s nickname[%s], for %s\n", phone, nickname, err.Error())
		return err
	}
	log.Printf("Set user[%s] nickname[%s]\n", phone, nickname)
	return nil
}

func UpdatePassword(phone, newPassword string) error {
	stmt, err := db.Prepare("UPDATE users set password=? where phone=?")
	checkErr(err)
	_, err = stmt.Exec(newPassword, phone)
	if err != nil {
		log.Printf("Cannot change user[%s]'s password to new password[%s], for %s\n", phone, newPassword, err.Error())
		return err
	}
	log.Printf("Change user[%s]'s password\n", newPassword)
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
			log.Printf("Find user[%s] failed\n, for %s", phone, err.Error())
			return [2]string{}, err
		}
		return [2]string{password, nullToEmpty(nickname)}, nil
	}
	log.Printf("Cannot find user[%s]\n", phone)
	return [2]string{"", ""}, nil
}

func SetKeyExpire(k string, expSec int) error {
	_, err := conn.Do("Expire", k, expSec)
	if err != nil {
		log.Printf("Cannot set(k=%s) expire time(t=%d) on redis, for %s\n", k, expSec, err.Error())
	}
	return err
}

func SetKeyValue(k, v string, expSec... int) error {
	_, err := conn.Do("set", k, v)
	if err != nil {
		log.Printf("Cannot set(k=%s, v=%s) on redis, for %s\n", k, v, err.Error())
	}
	if len(expSec) > 0 {
		err = SetKeyExpire(k, expSec[0])
	}
	return err
}

func SetKeyByteValue(k string, v []byte, expSec... int) error {
	_, err := conn.Do("set", k, v)
	if err != nil {
		log.Printf("Cannot set with byte value(k=%s, v=%v) on redis, for %s\n", k, v, err.Error())
	}
	if len(expSec) > 0 {
		err = SetKeyExpire(k, expSec[0])
	}
	return err
}

func GetKeyValue(k string) (string, error) {
	rep, err := conn.Do("get", k)
	if err != nil {
		log.Printf("Connot get value of k=%s, for %s\n", k, err.Error())
		return "", err
	}
	if rep == nil {
		return "", err
	}
	return string(rep.([]uint8)), err
}

func GetKeyByteValue(k string) ([]byte, error) {
	rep, err := conn.Do("get", k)
	if err != nil {
		log.Printf("Connot get value of k=%s, for %s\n", k, err.Error())
		return nil, err
	}
	if rep == nil {
		return nil, err
	}
	return (rep.([]uint8)), err
}
