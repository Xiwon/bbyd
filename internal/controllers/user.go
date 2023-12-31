package controllers

import (
	"time"
	"strconv"
	"net/http"

	"bbyd/internal/controllers/auth"
	"bbyd/internal/model"
	"bbyd/internal/shared/config"
	resp "bbyd/pkg/utils/response"

	"github.com/labstack/echo/v4"
	"github.com/go-playground/validator/v10"
)

type UserProfile struct {
	Uid      uint
	Username string
	Email    string
	Auth     string
}
var validate *validator.Validate = validator.New()

type loginRqst struct {
	Name   string `json:"name"   form:"name"   query:"name"`
	Passwd string `json:"passwd" form:"passwd" query:"passwd"`
}
type registerRqst struct {
	Name   string `json:"name"   form:"name"   query:"name"   validate:"required"`
	Passwd string `json:"passwd" form:"passwd" query:"passwd" validate:"required"`
	Repeat string `json:"repeat" form:"repeat" query:"repeat" validate:"required,eqfield=Passwd"`
	Email  string `json:"email"  form:"email"  query:"email"`
}
type setinfoRqst struct {
	Passwd string `json:"passwd" form:"passwd" query:"passwd"`
	Repeat string `json:"repeat" form:"repeat" query:"repeat"`
	Email  string `json:"email"  form:"email"  query:"email"`
	Auth   string `json:"auth"   form:"auth"   query:"auth"`
}

type loginResp struct {
	Token                 string `json:"token"`
	Token_expiration_time int64  `json:"token_expiration_time"`
}
type loginByEmailRqst struct {
	Name string `json:"name" form:"name" query:"name"`
}
type loginByCodeRqst struct {
	Code string `json:"code" form:"code" query:"code"`
}
type logoutResp loginResp

func UserModelToUserProfile(usr model.UserModel) UserProfile {
	return UserProfile{
		Uid:      usr.ID,
		Username: usr.Username,
		Email:    usr.Email,
		Auth:     usr.Auth,
	}
}

// usr := GetProfile(c)
// assume middleware has get model.UserModel from database
// & set `token_usr` to a controllers.UserProfile object
func GetProfile(c echo.Context) UserProfile {
	usr := c.Get("token_usr")
	return *usr.(*UserProfile)
}

// GET /user/:name
// authorized
func UserIndexHandler(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	usr := GetProfile(c)
	name := c.Param("name")
	if usr.Auth != config.Configs.Constants.AdminAuthname && usr.Username != name {
		return c.BYResponse(http.StatusBadRequest, "you are not an admin", nil)
	}
	mod, err := model.GetUsrByName(name)
	if err != nil {
		return c.BYResponse(http.StatusBadRequest, "database error", err.Error())
	}
	return c.BYResponse(http.StatusOK, "Welcome! Have a nice day",
		UserModelToUserProfile(mod))
}

// POST /user
// unauthorized
func RegisterHandler(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	req := new(registerRqst)
	err := c.Bind(req)
	if err != nil {
		return err
	}

	err = validate.Struct(req)
	if err != nil {
		return c.BYResponse(http.StatusBadRequest, "invalid form", err.Error())
	}

	err = model.TryRegister(req.Name, req.Passwd, req.Email)
	if err != nil {
		return c.BYResponse(http.StatusBadRequest, "database error", err.Error())
	}
	return c.BYResponse(http.StatusOK, "successfully registered", nil)
}

// PUT /user/:name
// authorized
func SetinfoHandler(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	req := new(setinfoRqst)
	name := c.Param("name")
	err := c.Bind(req)
	if err != nil {
		return err
	}
	usr := GetProfile(c)
	if req.Passwd != "" || req.Repeat != "" || req.Email != "" {
		if usr.Auth != config.Configs.Constants.AdminAuthname && usr.Username != name {
			return c.BYResponse(http.StatusBadRequest, "you are not an admin", nil)
		}
		if name == config.Configs.Constants.RootName && usr.Username != name {
			return c.BYResponse(http.StatusBadRequest, "NOBODY can change root info except root", nil)
		}
		if req.Passwd != "" && req.Passwd != req.Repeat {
			return c.BYResponse(http.StatusBadRequest, "invaild password", nil)
		}
	}
	if req.Auth != "" {
		if usr.Auth != config.Configs.Constants.AdminAuthname {
			return c.BYResponse(http.StatusBadRequest, "you are not an admin", nil)
		}
		if name == config.Configs.Constants.RootName {
			return c.BYResponse(http.StatusBadRequest, "CANNOT change root auth", nil)
		}
	}

	msg, err := model.TryChangeInfo(name, req.Passwd, req.Email, req.Auth)
	if err != nil {
		return c.BYResponse(http.StatusBadRequest, msg, err.Error())
	}
	return c.BYResponse(http.StatusOK, msg, nil)
}

// DELETE /user/:name
// authorized
func DeleteHandler(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	name := c.Param("name")

	usr := GetProfile(c)
	if usr.Auth != config.Configs.Constants.AdminAuthname && usr.Username != name {
		return c.BYResponse(http.StatusBadRequest, "you are not an admin", nil)
	}
	if name == config.Configs.Constants.RootName {
		return c.BYResponse(http.StatusBadRequest, "CANNOT delete root", nil)
	}

	msg, err := model.TryDelete(name)
	if err != nil {
		return c.BYResponse(http.StatusBadRequest, msg, err.Error())
	}
	return c.BYResponse(http.StatusOK, msg, nil)
}

// GET /user/token
// unauthorized
func LoginHandler(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	req := new(loginRqst)
	err := c.Bind(req)
	if err != nil {
		return err
	}

	legal := false
	{ // password verify
		db_sec, err := model.GetSecretByName(req.Name)
		if err != nil {
			return c.BYResponse(http.StatusBadRequest, "database error", err.Error())
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

	token, expireAt, err := auth.GenerateToken(req.Name)
	if err != nil {
		return c.BYResponse(http.StatusInternalServerError, err.Error(), nil)
	}
	return c.BYResponse(http.StatusOK, "login user "+req.Name, loginResp{
		Token:                 token,
		Token_expiration_time: expireAt,
	})
}

// GET /user/token/email
// unauthorized
func LoginByEmailHandler(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	req := new(loginByEmailRqst)
	err := c.Bind(req)
	if err != nil {
		return err
	}

	mod, err := model.GetUsrByName(req.Name)
	if err != nil {
		return c.BYResponse(http.StatusBadRequest, "database error", err.Error())
	}
	email := mod.Email

	las, err := model.GetEmailLastSendTime(email)
	if err == nil && las != nil {
		timestamp, err := strconv.ParseInt(las.(string), 10, 64)
		if err != nil {
			return c.BYResponse(http.StatusInternalServerError, "ParseInt failed", err.Error())
		}
		if time.Now().Unix() - timestamp < 60 {
			return c.BYResponse(http.StatusBadRequest, "too frequent requests", nil)
		}
	}

	code := auth.GenerateVerificationCode()
	err = auth.SendVerificationCodeEmail(email, code, req.Name)
	if err != nil {
		return c.BYResponse(http.StatusInternalServerError, "send verification email failed", err.Error())
	}
	err = model.UpdateCodeSendRecord(email, code, req.Name)
	if err != nil {
		return c.BYResponse(http.StatusInternalServerError, "update verification code failed", err.Error())
	}
	return c.BYResponse(http.StatusOK, "verification code has been sent to " + email, nil)
}

// GET /user/token/vcode
// unauthorized
func LoginByCodeHandler(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	req := new(loginByCodeRqst)
	err := c.Bind(req)
	if err != nil {
		return err
	}

	user, err := model.VerifyUsrByCode(req.Code)
	if err != nil {
		return c.BYResponse(http.StatusBadRequest, "database error", err.Error())
	}

	token, expireAt, err := auth.GenerateToken(user)
	if err != nil {
		return c.BYResponse(http.StatusInternalServerError, err.Error(), nil)
	}
	return c.BYResponse(http.StatusOK, "login user "+user, loginResp{
		Token:                 token,
		Token_expiration_time: expireAt,
	})
}

// DELETE /user/token
// authorized
func LogoutHandler(cc echo.Context) error {
	c := cc.(*resp.ResponseContext)
	usr := GetProfile(c)
	// fake logout
	return c.BYResponse(http.StatusOK, "logout from user "+usr.Username, logoutResp{
		Token:                 "",
		Token_expiration_time: 0,
	})
	// @todo: create token blacklist to immediately dispose invalid tokens
}
