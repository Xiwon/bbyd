package middleware

import(
	resp "bbyd/utils/response"
	"github.com/labstack/echo/v4"
	"bbyd/controllers/auth"
	"bbyd/utils/mark"
	contro "bbyd/controllers"
	"bbyd/model"
	"github.com/dgrijalva/jwt-go"
	// "go.uber.org/zap"
)

func TokenVerify(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		r := resp.New(c, "TokenVerify")
		raw, err := c.Cookie("token")
		if err != nil {
			return err
		}
		if raw.Value == "" {
			return r.Re(mark.BadRqst, "not login", nil)
		}

		token, err := jwt.ParseWithClaims(string(raw.Value), 
			&auth.TokenClaims{}, func(token_ *jwt.Token) (interface{}, error) {
			return auth.GetSkey(), nil
		})
		if err != nil {
			return r.Re(mark.BadRqst, "expired token", nil)
		}

		name := string(token.Claims.(*auth.TokenClaims).Username)
		usr, err := model.GetUsrByName(name) // get raw data from database
		if err != nil {
			return r.Re(mark.BadRqst, "no such user", nil)
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
