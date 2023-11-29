package model

import (
	"errors"
	"fmt"

	"bbyd/internal/shared/config"
	"bbyd/internal/controllers/auth"
	"bbyd/pkg/utils/logs"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var db *gorm.DB

type UserModel struct {
	gorm.Model
	Username string
	Secret   string
	Email    string
	Auth     string
}

func AutoMigrate(d config.Database) {
	db.AutoMigrate(&UserModel{})

	// pre-register root admin
	if d.PreRegisterRoot {
		err := TryRegister(config.Configs.Constants.RootName, config.Configs.Constants.RootDefaultPasswd, "")
		if err != nil {
			logs.Warn("root register failed", zap.Error(err))
		}
		_, err = TryChangeInfo(config.Configs.Constants.RootName, "", "", config.Configs.Constants.AdminAuthname)
		if err != nil {
			logs.Warn("root auth change failed", zap.Error(err))
		}
	}

	var usrdatas []UserModel
	db.Find(&usrdatas)
	fmt.Println("database status:\n", usrdatas)
}

// usr, err := model.GetUsrByName(name)
func GetUsrByName(name string) (UserModel, error) {
	var user UserModel
	result := db.First(&user, "username = ?", name)
	if result.RowsAffected == 0 {
		return UserModel{}, errors.New("not found")
	}
	return user, nil
}

// db_sec, err := GetSecretByName(req.Name)
func GetSecretByName(name string) (string, error) {
	var user UserModel
	result := db.First(&user, "username = ?", name)
	if result.RowsAffected == 0 {
		return "", errors.New("not found")
	}
	return user.Secret, nil
}

// err = TryRegister(req.Name, req.Passwd, req.Email)
func TryRegister(name string, passwd string, email string) error {
	var user UserModel
	result := db.First(&user, "username = ?", name)
	if result.RowsAffected > 0 {
		return errors.New("exist user")
	}
	user = UserModel{
		Username: name,
		Secret:   auth.GenerateSecret(passwd, auth.GenerateSalt()),
		Email:    email,
		Auth:     config.Configs.Constants.DefaultAuthname,
	}
	db.Create(&user)
	return nil
}

// msg, err := TryChangeInfo(req.Name, req.Passwd, req.Email)
func TryChangeInfo(name string, passwd string, email string, authoriz string) (string, error) {
	var user UserModel
	result := db.First(&user, "username = ?", name)
	if result.RowsAffected == 0 {
		return "user " + name + " not found",
			errors.New("user not found")
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
		db.Save(&user)
	}

	return msg, nil
}

// msg, err := TryDelete(req.Name)
func TryDelete(name string) (string, error) {
	var user UserModel
	result := db.First(&user, "username = ?", name)
	if result.RowsAffected == 0 {
		return "user " + name + " not found",
			errors.New("user not found")
	}
	db.Delete(&user)
	return "delete user " + name, nil
}
