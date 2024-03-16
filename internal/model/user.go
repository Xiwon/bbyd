package model

import (
	"bbyd/internal/shared/config"
	"bbyd/internal/controllers/auth"
	// "bbyd/pkg/utils/logs"

	// "go.uber.org/zap"
	"gorm.io/gorm"
	// "github.com/garyburd/redigo/redis"
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