package controllers

import(
	// "fmt"
	"time"
	"net/http"

	"bbyd/internal/model"
	"bbyd/internal/controllers/auth"
	resp "bbyd/pkg/utils/response"
	
	"github.com/labstack/echo/v4"
)

type UserProfile struct {
	Uid uint
	Username string
	Email string
	Auth string
}

type loginRqst struct {
	Name   string `json:"name"   form:"name"   query:"name"`
	Passwd string `json:"passwd" form:"passwd" query:"passwd"`
}
type registerRqst struct {
	Name   string `json:"name"   form:"name"   query:"name"`
	Passwd string `json:"passwd" form:"passwd" query:"passwd"`
	Repeat string `json:"repeat" form:"repeat" query:"repeat"`
	Email  string `json:"email"  form:"email"  query:"email"`
}
type setauthRqst struct {
	Name   string `json:"name"   form:"name"   query:"name"`
	To     string `json:"to"     form:"to"     query:"to"`
}
type setinfoRqst registerRqst
type deleteRqst struct {
	Name   string `json:"name"   form:"name"   query:"name"`
}

// usr := GetProfile(c)
// assume middleware has get model.UserModel from database 
// & set `token_usr` to a controllers.UserProfile object
func GetProfile(c echo.Context) UserProfile {
	usr := c.Get("token_usr")
	return *usr.(*UserProfile)
}

// assume user has been verified
func UserGET(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	usr := GetProfile(c)
	return c.BYResponse(http.StatusOK, "Welcome! Have a nice day", usr)
}

func LoginPOST(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	req := new(loginRqst)
	err := c.Bind(req)
	if err != nil {
		return err
	}

	legal := false
	{	// password verify
		db_sec, err := model.GetSecretByName(req.Name)
		if err != nil {
			return c.BYResponse(http.StatusBadRequest, "user not found", nil)
		}
		salt := auth.GetSaltFromSecret(db_sec)
		sec := auth.GenerateSecret(req.Passwd, salt)
		// get salt & generate secret
		if sec == db_sec {
			legal = true
		}
	}
	if !legal {
		return c.BYResponse(http.StatusBadRequest, "wrong password", nil)
	}

	cookie := &http.Cookie{
		Name: "token",
		Value: auth.GenerateToken(req.Name),
		HttpOnly: true,
	}
	c.SetCookie(cookie)
	return c.BYResponse(http.StatusOK, "login user " + req.Name, nil)
}

func RegisterPOST(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	req := new(registerRqst)
	err := c.Bind(req)
	if err != nil {
		return err
	}

	if req.Name == "" {
		return c.BYResponse(http.StatusBadRequest, "empty username is not allowed", nil)
	}
	if req.Passwd == "" {
		return c.BYResponse(http.StatusBadRequest, "empty password is not allowed", nil)
	}
	if req.Passwd != req.Repeat {
		return c.BYResponse(http.StatusBadRequest, "repeat password differs from the first input", nil)
	}

	err = model.TryRegister(req.Name, req.Passwd, req.Email)
	if err != nil {
		return c.BYResponse(http.StatusBadRequest, "existing name", nil)
	}
	return c.BYResponse(http.StatusOK, "successfully registered", nil)
}

// assume user has been verified
func LogoutPOST(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	usr := GetProfile(c)
	cookie := &http.Cookie{
		Name: "token",
		Value: "",
		Expires: time.Now().Add(-1e9),
		MaxAge: -1,
	}
	c.SetCookie(cookie)

	return c.BYResponse(http.StatusOK, "logout from user " + usr.Username, nil)
}

// assume user has been verified
func SetauthPOST(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	req := new(setauthRqst)
	err := c.Bind(req)
	if err != nil {
		return err
	}
	usr := GetProfile(c)
	if usr.Auth != "admin" {
		return c.BYResponse(http.StatusBadRequest, "you are not an admin", nil)
	}
	if req.Name == "root" {
		return c.BYResponse(http.StatusBadRequest, "CANNOT change root auth", nil)
	}

	err = model.TryChangeAuth(req.Name, req.To)
	if err != nil {
		return c.BYResponse(http.StatusBadRequest, "user " + req.Name + " not found", nil)
	}
	return c.BYResponse(http.StatusOK, "you've changed user " + req.Name + "'s Auth to " + req.To, nil)
}

// assume user has been verified
func SetinfoPOST(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	req := new(setinfoRqst)
	err := c.Bind(req)
	if err != nil {
		return err
	}
	usr := GetProfile(c)
	if usr.Auth != "admin" && usr.Username != req.Name {
		return c.BYResponse(http.StatusBadRequest, "you are not an admin", nil)
	}
	if req.Name == "root" && usr.Username != "root" {
		return c.BYResponse(http.StatusBadRequest, "NOBODY can change root info except root", nil)
	}

	msg, err := model.TryChangeInfo(req.Name, req.Passwd, req.Email)
	if err != nil {
		return c.BYResponse(http.StatusBadRequest, msg, nil)
	}
	return c.BYResponse(http.StatusOK, msg, nil)
}

// assume user has been verified
func DeletePOST(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	req := new(deleteRqst)
	err := c.Bind(req)
	if err != nil {
		return err
	}

	usr := GetProfile(c)
	if usr.Auth != "admin" && usr.Username != req.Name {
		return c.BYResponse(http.StatusBadRequest, "you are not an admin", nil)
	}
	if req.Name == "root" {
		return c.BYResponse(http.StatusBadRequest, "CANNOT delete root", nil)
	}

	msg, err := model.TryDelete(req.Name)
	if err != nil {
		return c.BYResponse(http.StatusBadRequest, msg, nil)
	}
	return c.BYResponse(http.StatusOK, msg, nil)
}