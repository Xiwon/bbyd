package model

import (
	"time"
	"errors"
	"fmt"
	"strconv"

	"bbyd/internal/shared/config"
	// "bbyd/internal/controllers/auth"
	// "bbyd/pkg/utils/logs"

	// "go.uber.org/zap"
	"gorm.io/gorm"
	"github.com/garyburd/redigo/redis"
)

var (
	DBInternalError = errors.New("Database Internal Error")
	DBExistError    = errors.New("Database Exist Error")
	DBNotFoundError = errors.New("Database Not Found")
	RedisInternalError = errors.New("Redis Internal Error")
	RedisExistError = errors.New("Redis Exist Error")
	RedisNotFoundError = errors.New("Redis Not Found")
)

var db *gorm.DB           // postgresql
var redisConn redis.Conn // redis

func AutoMigrate(d config.Database) error {
	err := db.AutoMigrate(
		&UserModel{},
		&NodeModel{},
		&PostModel{},
	)
	if err != nil {
		return err
	}

	// pre-register root admin
	if d.PreRegisterRoot {
		err := TryRegister(config.Configs.Constants.RootName, config.Configs.Constants.RootDefaultPasswd, "")
		if err != nil {
			fmt.Println("root register failed")
		}
		_, err = TryChangeInfo(config.Configs.Constants.RootName, "", "", config.Configs.Constants.AdminAuthname)
		if err != nil {
			fmt.Println("root auth change failed")
		}
	}

	var usrdatas []UserModel
	err = db.Find(&usrdatas).Error
	if err != nil {
		return DBInternalError
	}
	fmt.Println("[TABLE user_models]:")
	for _, v := range usrdatas {
		fmt.Printf("{ Username:%s, Email:%s, Auth:%s }\n", 
			v.Username, v.Email, v.Auth)
	}

	return nil
}

func VerifyUsrByCode(code string) (string, error) {
	user, err := redisConn.Do("get", "codeUser:" + code)
	if err != nil {
		return "", RedisInternalError
	}
	if user == nil { // not found
		return "", RedisNotFoundError
	}

	sendTime, err := redisConn.Do("get", "codeSendTime:" + code)
	if err != nil {
		return "", RedisInternalError
	}
	// assert sendTime != nil
	timeStamp, err := strconv.ParseInt(sendTime.(string), 10, 64)
	expired := false
	if time.Now().Unix() - timeStamp > 
		60 * int64(config.Configs.SmtpConfig.CodeExpirationMinute) { // verification code expired 
		expired = true
	}

	_, err = redisConn.Do("del", "codeUser:" + code, "codeSendTime:" + code)
	if err != nil {
		return "", RedisInternalError
	}

	if (expired) {
		return "", RedisNotFoundError
	}

	return user.(string), nil
}