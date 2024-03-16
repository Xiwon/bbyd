package model

import (
	"bbyd/internal/shared/config"
	"bbyd/internal/shared/generator"

	// "bbyd/internal/controllers/auth"
	// "bbyd/pkg/utils/logs"

	// "go.uber.org/zap"
	"gorm.io/gorm"
	// "github.com/gomodule/redigo/redis"
)

type UserModel struct {
	gorm.Model
	Username string
	Secret   string
	Email    string
	Auth     string
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
		user.Secret = generator.GenerateSecret(passwd, generator.GenerateSalt())
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
		Secret:   generator.GenerateSecret(passwd, generator.GenerateSalt()),
		Email:    email,
		Auth:     config.Configs.Constants.DefaultAuthname,
	}
	err := db.Create(&user).Error
	if err != nil {
		return DBInternalError
	}
	return nil
}

// err = model.SetEmailVerifyCode(key, name, conf.CodeExpirationMinute * 60)
func SetEmailVerifyCode(key, name string, ttlSeconds int) error {
	res, err := redisConn.Do("exists", key)
	if err != nil {
		return err
	}
	if res.(int64) != 0 {
		return RedisExistError
	}

	_, err = redisConn.Do("set", key, name)
	if err != nil {
		return err
	}
	_, err = redisConn.Do("expire", key, ttlSeconds)
	if err != nil {
		return err
	}

	return nil
}

// res, err := model.CodeNameGet(code)
func CodeNameGet(key string) (interface{}, error) {
	return redisConn.Do("get", key)
}

// err := model.SetExpiredToken(rawToken)
func SetExpiredToken(tokenKey, name string) error {
	_, err := redisConn.Do("set", tokenKey, name)
	if err != nil {
		return err
	}
	_, err = redisConn.Do("expire", tokenKey, config.Configs.Constants.TokenExpirationMinute*60)
	if err != nil {
		return err
	}

	return nil
}

func CheckDeprecatedToken(key string) bool {
	res, err := redisConn.Do("exists", key)
	if err != nil {
		return true // login not allowed when internal error occurred
	}
	if res.(int64) == 0 {
		return false
	}
	return true
}
