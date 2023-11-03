package main

import(
	"fmt"
	"time"
	"net/http"
	"crypto/sha256"

	"bbyd/db"
	"bbyd/jwt"

	"github.com/labstack/echo/v4"
)

type msg_json struct {
	Msg string
}
type profile_json struct {
	Msg string
	Profile db.UserProfile
}
type login_rqst struct {
	Name   string `json:"name"   form:"name"   query:"name"`
	Passwd string `json:"passwd" form:"passwd" query:"passwd"`
}
type register_rqst struct {
	Name   string `json:"name"   form:"name"   query:"name"`
	Passwd string `json:"passwd" form:"passwd" query:"passwd"`
	Repeat string `json:"repeat" form:"repeat" query:"repeat"`
	Email  string `json:"email"  form:"email"  query:"email"`
}
type setauth_rqst struct {
	Name   string `json:"name"   form:"name"   query:"name"`
	To     string `json:"to"     form:"to"     query:"to"`
}
type delete_rqst struct {
	Name   string `json:"name"   form:"name"   query:"name"`
}

func get_token_usr(c echo.Context) (token_usr string) {
	if res := c.Get("token_usr"); res != nil {
		token_usr = res.(*db.UserProfile).Username
	}
	return
}

func registerRoutes(e *echo.Echo) {
	e.Any("/", func(c echo.Context) error {
		// access main page
		token_usr := get_token_usr(c)

		if token_usr != "" {
			usr, ok := db.GetUsrByName(token_usr)
			if ok == 0 {
				res := &msg_json{ Msg: "(/)[failed] deleted user" }
				return c.JSONPretty(http.StatusBadRequest, res, "  ")
			} else {
				res := &profile_json{
					Msg: "Welcome! Have a nice day",
					Profile: usr,
				}
				return c.JSONPretty(http.StatusOK, res, "  ")
			}
		} else {
			res := &msg_json{ Msg: "(/)[failed] you did't login, or your token had expired" }
			return c.JSONPretty(http.StatusUnauthorized, res, "  ")
		}
	})

	e.Any("/login", func(c echo.Context) error {
		req := new(login_rqst)
		if err := c.Bind(req); err != nil {
			return err
		}
		token_usr := get_token_usr(c)

		db_sec, ok := db.GetSecretByName(req.Name)
		sec_sum := sha256.Sum256([]byte(req.Passwd))
		sec := fmt.Sprintf("%x", sec_sum)

		if ok == 1 && db_sec == sec {
			// right
			cookie := &http.Cookie{
				Name: "token",
				Value: jwt.GenerateToken(req.Name),
				HttpOnly: true,
			}
			c.SetCookie(cookie)
			fmt.Println("route.go: set token:", cookie.Value)

			msg := "(/login)[success] "
			if token_usr == "" {
				msg = msg + "you've login in user " + req.Name + " from anonymous"
			} else if token_usr != req.Name {
				msg = msg + "you've switch to user " + req.Name + " from " + token_usr
			} else if token_usr == req.Name {
				msg = msg + "you've flashed your cookie"
			}

			res := &msg_json{ Msg: msg }
			return c.JSONPretty(http.StatusOK, res, "  ")
		} else {
			res := &msg_json{ Msg: "(/login)[failed] wrong name or passwd" }
			return c.JSONPretty(http.StatusForbidden, res, "  ")
		}
	})

	e.Any("/register", func(c echo.Context) error {
		req := new(register_rqst)
		if err := c.Bind(req); err != nil {
			return err
		}
		if req.Name == "" {
			return c.JSONPretty(http.StatusBadRequest, &msg_json{
				Msg: "(/register)[failed] empty username is not allowed",
			}, "  ")
		}
		if req.Passwd == "" {
			return c.JSONPretty(http.StatusBadRequest, &msg_json{
				Msg: "(/register)[failed] empty password is not allowed",
			}, "  ")
		}
		if req.Passwd != req.Repeat {
			return c.JSONPretty(http.StatusBadRequest, &msg_json{
				Msg: "(/register)[failed] repeat password differs from the first input",
			}, "  ")
		}

		ok := db.TryRegister(req.Name, req.Passwd, req.Email)
		if ok == 1 {
			res := &msg_json{ Msg: "(/register)[success] you've successfully registered" }
			return c.JSONPretty(http.StatusOK, res, "  ")
		} else {
			res := &msg_json{ Msg: "(/register)[failed] existing name" }
			return c.JSONPretty(http.StatusBadRequest, res, "  ")
		}
	})

	e.Any("/logout", func(c echo.Context) error {
		token_usr := get_token_usr(c)

		if token_usr == "" {
			res := &msg_json{ Msg: "(/logout)[failed] you didn't login" }
			return c.JSONPretty(http.StatusBadRequest, res, "  ")
		} else {
			cookie := &http.Cookie{
				Name: "token",
				Value: "",
				Expires: time.Now().Add(-1e9),
				MaxAge: -1,
			}
			c.SetCookie(cookie)

			res := &msg_json{ Msg: "(/logout)[success] logout from user " + token_usr }
			return c.JSONPretty(http.StatusOK, res, "  ")
		}
	})

	e.Any("/setauth", func(c echo.Context) error {
		token_usr := get_token_usr(c)
		if token_usr == "" {
			res := &msg_json{ Msg: "(/setauth)[failed] you didn't login" }
			return c.JSONPretty(http.StatusBadRequest, res, "  ")
		}

		user, ok := db.GetUsrByName(token_usr)
		if ok == 1 && user.Auth == "admin" { // auth check
			req := new(setauth_rqst)
			if err := c.Bind(req); err != nil {
				return err
			}

			if req.Name == "" {
				return c.JSONPretty(http.StatusBadRequest, &msg_json{
					Msg: "(/setauth)[failed] empty target user",
				}, "  ")
			}
			if req.Name == "root" {
				return c.JSONPretty(http.StatusBadRequest, &msg_json{
					Msg: "(/setauth)[failed] CANNOT change root Auth",
				}, "  ")
			}
			if req.To == "" {
				return c.JSONPretty(http.StatusBadRequest, &msg_json{
					Msg: "(/setauth)[failed] empty user auth",
				}, "  ")
			}

			chgok := db.TryChangeAuth(req.Name, req.To)
			if chgok == 1 {
				res := &msg_json{ Msg: "(/setauth)[success] you've changed user " + req.Name + "'s Auth to " + req.To }
				return c.JSONPretty(http.StatusOK, res, "  ")
			} else {
				res := &msg_json{ Msg: "(/setauth)[failed] user " + req.Name + " didn't found" }
				return c.JSONPretty(http.StatusBadRequest, res, "  ")
			}
		} else {
			res := &msg_json{ Msg: "(/setauth)[failed] you are not an admin" }
			return c.JSONPretty(http.StatusBadRequest, res, "  ")
		}
	})

	e.Any("/setinfo", func(c echo.Context) error {
		token_usr := get_token_usr(c)
		if token_usr == "" {
			res := &msg_json{ Msg: "(/setinfo)[failed] you didn't login" }
			return c.JSONPretty(http.StatusBadRequest, res, "  ")
		}

		req := new(register_rqst)
		if err := c.Bind(req); err != nil {
			return err
		}

		user, ok := db.GetUsrByName(token_usr)
		if ok == 1 && (user.Auth == "admin" || 
			user.Username == req.Name) { // auth check

			if req.Name == "root" && user.Username != "root" {
				return c.JSONPretty(http.StatusBadRequest, &msg_json{
					Msg: "(/setinfo)[failed] NOBODY can change root info except root",
				}, "  ")
			}

			chgmsg, chgok := db.TryChangeInfo(req.Name, req.Passwd, req.Email)
			if chgok == 1 {
				res := &msg_json{ Msg: "(/setinfo)[success] " + chgmsg }
				return c.JSONPretty(http.StatusOK, res, "  ")
			} else {
				res := &msg_json{ Msg: "(/setinfo)[failed] " + chgmsg }
				return c.JSONPretty(http.StatusBadRequest, res, "  ")
			}
		} else {
			res := &msg_json{ Msg: "(/setinfo)[failed] you are not an admin" }
			return c.JSONPretty(http.StatusBadRequest, res, "  ")
		}
	})

	e.Any("/delete", func(c echo.Context) error {
		token_usr := get_token_usr(c)
		if token_usr == "" {
			res := &msg_json{ Msg: "(/delete)[failed] you didn't login" }
			return c.JSONPretty(http.StatusBadRequest, res, "  ")
		}

		req := new(delete_rqst)
		if err := c.Bind(req); err != nil {
			return err
		}

		user, ok := db.GetUsrByName(token_usr)
		if ok == 1 && (user.Auth == "admin" || 
			user.Username == req.Name) { // auth check

			if req.Name == "root" {
				return c.JSONPretty(http.StatusBadRequest, &msg_json{
					Msg: "(/delete)[failed] CANNOT delete root user",
				}, "  ")
			}
			if req.Name == "" {
				return c.JSONPretty(http.StatusBadRequest, &msg_json{
					Msg: "(/delete)[failed] empty target user",
				}, "  ")
			}

			delmsg, delok := db.TryDeleteByName(req.Name)
			if delok == 1 {
				res := &msg_json{ Msg: "(/delete)[success] " + delmsg }
				return c.JSONPretty(http.StatusOK, res, "  ")
			} else {
				res := &msg_json{ Msg: "(/delete)[failed] " + delmsg }
				return c.JSONPretty(http.StatusBadRequest, res, "  ")
			}
		} else {
			res := &msg_json{ Msg: "(/delete)[failed] you are not an admin" }
			return c.JSONPretty(http.StatusBadRequest, res, "  ")
		}
	})
}