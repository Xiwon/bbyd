package controllers

import(
	"time"
	"net/http"

	"bbyd/utils/mark"
	"bbyd/model"
	"bbyd/controllers/auth"
	resp "bbyd/utils/response"
	
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
	return UserProfile{
		Uid: usr.(*UserProfile).Uid,
		Username: usr.(*UserProfile).Username,
		Email: usr.(*UserProfile).Email,
		Auth: usr.(*UserProfile).Auth,
	}
}

// assume user has been verified
func UserGET(c echo.Context) error {
	r := resp.New(c, "/user")
	usr := GetProfile(c)
	return r.Re(mark.OK, "Welcome! Have a nice day", usr)
}

func LoginPOST(c echo.Context) error {
	r := resp.New(c, "/login")
	req := new(loginRqst)
	err := c.Bind(req)
	if err != nil {
		return err
	}

	legal := false
	{	// password verify
		db_sec, err := model.GetSecretByName(req.Name)
		if err != nil {
			return r.Re(mark.BadRqst, "user not found", nil)
		}
		salt := auth.GetSaltFromSecret(db_sec)
		sec := auth.GenerateSecret(req.Passwd, salt)
		// get salt & generate secret
		if sec == db_sec {
			legal = true
		}
	}
	if !legal {
		return r.Re(mark.BadRqst, "wrong password", nil)
	}

	cookie := &http.Cookie{
		Name: "token",
		Value: auth.GenerateToken(req.Name),
		HttpOnly: true,
	}
	c.SetCookie(cookie)
	return r.Re(mark.OK, "login user " + req.Name, nil)
}

func RegisterPOST(c echo.Context) error {
	r := resp.New(c, "/register")

	req := new(registerRqst)
	err := c.Bind(req)
	if err != nil {
		return err
	}

	if req.Name == "" {
		return r.Re(mark.BadRqst, "empty username is not allowed", nil)
	}
	if req.Passwd == "" {
		return r.Re(mark.BadRqst, "empty password is not allowed", nil)
	}
	if req.Passwd != req.Repeat {
		return r.Re(mark.BadRqst, "repeat password differs from the first input", nil)
	}

	err = model.TryRegister(req.Name, req.Passwd, req.Email)
	if err != nil {
		return r.Re(mark.BadRqst, "existing name", nil)
	}
	return r.Re(mark.OK, "successfully registered", nil)
}

// assume user has been verified
func LogoutPOST(c echo.Context) error {
	r := resp.New(c, "/logout")
	usr := GetProfile(c)
	cookie := &http.Cookie{
		Name: "token",
		Value: "",
		Expires: time.Now().Add(-1e9),
		MaxAge: -1,
	}
	c.SetCookie(cookie)

	return r.Re(mark.OK, "logout from user " + usr.Username, nil)
}

// assume user has been verified
func SetauthPOST(c echo.Context) error {
	r := resp.New(c, "/setauth")
	req := new(setauthRqst)
	err := c.Bind(req)
	if err != nil {
		return err
	}
	usr := GetProfile(c)
	if usr.Auth != "admin" {
		return r.Re(mark.BadRqst, "you are not an admin", nil)
	}
	if req.Name == "root" {
		return r.Re(mark.BadRqst, "CANNOT change root auth", nil)
	}

	err = model.TryChangeAuth(req.Name, req.To)
	if err != nil {
		return r.Re(mark.BadRqst, "user " + req.Name + " not found", nil)
	}
	return r.Re(mark.OK, "you've changed user " + req.Name + "'s Auth to " + req.To, nil)
}

// assume user has been verified
func SetinfoPOST(c echo.Context) error {
	r := resp.New(c, "/setinfo")
	req := new(setinfoRqst)
	err := c.Bind(req)
	if err != nil {
		return err
	}
	usr := GetProfile(c)
	if usr.Auth != "admin" && usr.Username != req.Name {
		return r.Re(mark.BadRqst, "you are not an admin", nil)
	}
	if req.Name == "root" && usr.Username != "root" {
		return r.Re(mark.BadRqst, "NOBODY can change root info except root", nil)
	}

	msg, err := model.TryChangeInfo(req.Name, req.Passwd, req.Email)
	if err != nil {
		return r.Re(mark.BadRqst, msg, nil)
	}
	return r.Re(mark.OK, msg, nil)
}

// assume user has been verified
func DeletePOST(c echo.Context) error {
	r := resp.New(c, "/delete")
	req := new(deleteRqst)
	err := c.Bind(req)
	if err != nil {
		return err
	}

	usr := GetProfile(c)
	if usr.Auth != "admin" && usr.Username != req.Name {
		return r.Re(mark.BadRqst, "you are not an admin", nil)
	}
	if req.Name == "root" {
		return r.Re(mark.BadRqst, "CANNOT delete root", nil)
	}

	msg, err := model.TryDelete(req.Name)
	if err != nil {
		return r.Re(mark.BadRqst, msg, nil)
	}
	return r.Re(mark.OK, msg, nil)
}