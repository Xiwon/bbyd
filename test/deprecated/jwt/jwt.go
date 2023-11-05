package jwt

import(
	"fmt"
	"time"

	"bbyd/db"

	"github.com/labstack/echo/v4"
	d_jwt "github.com/dgrijalva/jwt-go"
)

type claims struct {
	Username string `json:"username"`
	d_jwt.StandardClaims
}

var jwt_skey = []byte("bbingyan_jwt_skey_58490998")

func AutoLogin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		raw, raw_err := c.Cookie("token")

		if raw_err == nil && raw.Value != "" {
			var s string = string(raw.Value)
			token, t_err := d_jwt.ParseWithClaims(s, 
				&claims{}, func(token_ *d_jwt.Token) (interface{}, error) {
				return jwt_skey, nil
			})

			if t_err == nil {
				// unexpired token
				var token_usr string = string(token.Claims.(*claims).Username)
				fmt.Println("jwt.go: middleware get token:", token_usr)
				c.Set("token_usr", &db.UserProfile{ Username: token_usr })
			} else {
				fmt.Println("jwt.go: get token but parse error")
			}
		}

		return next(c)
	}
}

func GenerateToken(name string) string {
	c := claims{
		Username: name,
		StandardClaims: d_jwt.StandardClaims{
        	NotBefore: time.Now().Unix() - 60,
        	ExpiresAt: time.Now().Unix() + 60,
        	Issuer: name,
		},
	}

	t := d_jwt.NewWithClaims(d_jwt.SigningMethodHS256, c)
	s, _ := t.SignedString(jwt_skey)
	return string(s)
}