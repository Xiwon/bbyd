package middleware

import (
	"net/http"

	contro "bbyd/internal/controllers"
	"bbyd/internal/controllers/auth"
	"bbyd/internal/model"
	resp "bbyd/pkg/utils/response"

	"github.com/labstack/echo/v4"
)

func TokenVerify(next echo.HandlerFunc) echo.HandlerFunc {
	return func(cc echo.Context) error {
		c := cc.(*resp.ResponseContext)
		claims, err := auth.GetClaimsFromHeader(c)
		// get token from header and parse it while checking validity
		if err != nil {
			return c.BYResponse(http.StatusBadRequest, err.Error(), nil)
		}

		mod, err := model.GetUsrByName(claims.Username) // get raw data from database
		if err != nil {
			return c.BYResponse(http.StatusBadRequest, "database error", err.Error())
		}
		usr := contro.UserModelToUserProfile(mod)
		c.Set("token_usr", &usr)
		c.Set("raw_token", &claims.RawToken)
		// set `token_usr` to contro.UserProfile object

		return next(c)
	}
}

// if access with token - verify token
// if not               - skip and access unauthorized
func CanHaveToken(next echo.HandlerFunc) echo.HandlerFunc {
	return func(cc echo.Context) error {
		c := cc.(*resp.ResponseContext)
		claims, err := auth.GetClaimsFromHeader(c)
		if err != nil {
			// access without authorization
			c.Set("token_usr", new(contro.UserProfile))
			c.Set("raw_token", new(string))
			return next(c)
		}

		mod, err := model.GetUsrByName(claims.Username) // get raw data from database
		if err != nil {
			return c.BYResponse(http.StatusBadRequest, "database error", err.Error())
		}
		usr := contro.UserModelToUserProfile(mod)
		c.Set("token_usr", &usr)
		c.Set("raw_token", &claims.RawToken)
		// set `token_usr` to contro.UserProfile object

		return next(c)
	}
}
