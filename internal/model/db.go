package model

import (
	"errors"
	"fmt"

	"bbyd/internal/controllers/auth"
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

func AutoMigrate() {
	db.AutoMigrate(&UserModel{})

	// pre-register root admin
	MustRegister("root", "123456", "")
	MustChangeAuth("root", "admin")

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
		Auth:     "normal",
	}
	db.Create(&user)
	return nil
}
func MustRegister(name string, passwd string, email string) {
	TryRegister(name, passwd, email)
}

// err = TryChangeAuth(req.Name, req.To)
func TryChangeAuth(name string, auth string) error {
	var user UserModel
	result := db.First(&user, "username = ?", name)
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	user.Auth = auth
	db.Save(&user)
	return nil
}
func MustChangeAuth(name string, auth string) {
	TryChangeAuth(name, auth)
}

// msg, err := TryChangeInfo(req.Name, req.Passwd, req.Email)
func TryChangeInfo(name string, passwd string, email string) (string, error) {
	var user UserModel
	result := db.First(&user, "username = ?", name)
	if result.RowsAffected == 0 {
		return "user " + name + " not found",
			errors.New("user not found")
	}

	var msg string
	if email != "" {
		user.Email = email
		msg += " email changed to " + email
	}
	if passwd != "" {
		user.Secret = auth.GenerateSecret(passwd, auth.GenerateSalt())
		msg += " passwd changed"
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
