package model

import (
	"time"
	"errors"
	"fmt"
	"strconv"

	"bbyd/internal/shared/config"
	"bbyd/internal/controllers/auth"
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

type UserModel struct {
	gorm.Model
	Username string
	Secret   string
	Email    string
	Auth     string
}

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

// usr, err := model.GetUsrByName(name)
func GetUsrByName(name string) (UserModel, error) {
	var user UserModel
	result := db.First(&user, "username = ?", name)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return UserModel{}, DBNotFoundError
		}
		return UserModel{}, DBInternalError
	}

	return user, nil
}

// db_sec, err := GetSecretByName(req.Name)
func GetSecretByName(name string) (string, error) {
	var user UserModel
	result := db.First(&user, "username = ?", name)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return "", DBNotFoundError
		}
		return "", DBInternalError
	}
	
	return user.Secret, nil
}

// err = TryRegister(req.Name, req.Passwd, req.Email)
func TryRegister(name string, passwd string, email string) error {
	var user UserModel
	result := db.First(&user, "username = ?", name)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return DBInternalError
	}
	if result.RowsAffected > 0 {
		return DBExistError
	}
	user = UserModel{
		Username: name,
		Secret:   auth.GenerateSecret(passwd, auth.GenerateSalt()),
		Email:    email,
		Auth:     config.Configs.Constants.DefaultAuthname,
	}
	err := db.Create(&user).Error
	if err != nil {
		return DBInternalError
	}
	return nil
}

// msg, err := TryChangeInfo(req.Name, req.Passwd, req.Email)
func TryChangeInfo(name string, passwd string, email string, authoriz string) (string, error) {
	var user UserModel
	result := db.First(&user, "username = ?", name)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return "", DBInternalError
	}
	if result.Error == gorm.ErrRecordNotFound {
		return "user " + name + " not found", DBNotFoundError
	}

	var msg string
	if passwd != "" {
		user.Secret = auth.GenerateSecret(passwd, auth.GenerateSalt())
		msg += " passwd changed"
	}
	if email != "" {
		user.Email = email
		msg += " email changed to " + email
	}
	if authoriz != "" {
		user.Auth = authoriz
		msg += " auth changed to " + authoriz
	}
	if msg == "" {
		msg = "nothing changed"
	} else {
		err := db.Save(&user).Error
		if err != nil {
			return "", DBInternalError
		}
	}

	return msg, nil
}

// msg, err := TryDelete(req.Name)
func TryDelete(name string) (string, error) {
	var user UserModel
	result := db.First(&user, "username = ?", name)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return "", DBInternalError
	}
	if result.Error == gorm.ErrRecordNotFound {
		return "user " + name + " not found", DBNotFoundError
	}
	err := db.Delete(&user).Error
	if err != nil {
		return "", DBInternalError
	}
	return "delete user " + name, nil
}

func GetEmailLastSendTime(email string) (interface{}, error) {
	las, err := redisConn.Do("get", "emailLastSendTime:" + email)
	if err != nil {
		return "", RedisInternalError
	}
	return las, nil
}

func UpdateCodeSendRecord(email string, code string, name string) error {
	_, err := redisConn.Do("set", "emailLastSendTime:" + email, fmt.Sprintf("%d", time.Now().Unix()))
	if err != nil {
		return RedisInternalError
	}

	_, err = redisConn.Do("set", "codeUser:" + code, name)
	if err != nil {
		return RedisInternalError
	}

	_, err = redisConn.Do("set", "codeSendTime:" + code, fmt.Sprintf("%d", time.Now().Unix()))
	if err != nil {
		return RedisInternalError
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