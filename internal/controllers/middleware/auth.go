package middleware

import(
	"net/http"
	
	"bbyd/internal/model"
	"bbyd/internal/controllers/auth"
	contro "bbyd/internal/controllers"
	resp "bbyd/pkg/utils/response"
	
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
)

func TokenVerify(next echo.HandlerFunc) echo.HandlerFunc {
	return func(cc echo.Context) error {
		c := cc.(*resp.ResponseContext)
		raw, err := c.Cookie("token")
		if err != nil {
			return c.BYResponse(http.StatusBadRequest, "cookie not found", nil)
		}
		if raw.Value == "" {
			return c.BYResponse(http.StatusBadRequest, "not login", nil)
		}

		token, err := jwt.ParseWithClaims(string(raw.Value), 
			&auth.TokenClaims{}, func(token_ *jwt.Token) (interface{}, error) {
			return auth.GetSkey(), nil
		})
		if err != nil {
			return c.BYResponse(http.StatusBadRequest, "expired token", nil)
		}

		name := string(token.Claims.(*auth.TokenClaims).Username)
		usr, err := model.GetUsrByName(name) // get raw data from database
		if err != nil {
			return c.BYResponse(http.StatusBadRequest, "no such user", nil)
		}
		c.Set("token_usr", &contro.UserProfile{
			Uid: usr.ID,
			Username: usr.Username,
			Email: usr.Email,
			Auth: usr.Auth,
		})
		// set `token_usr` to contro.UserProfile object

		return next(c)
	}
}
