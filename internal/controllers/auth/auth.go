package auth

import (
	"errors"

	"strconv"
	"strings"
	"time"

	"bbyd/internal/model"
	"bbyd/internal/shared/config"
	"bbyd/internal/shared/generator"
	"bbyd/pkg/utils/response"

	"github.com/dgrijalva/jwt-go"
	"github.com/gophish/gomail"
)

type Claims struct {
	Username string `json:"username"`
	RawToken string `json:"rawtoken"`
	jwt.StandardClaims
}

var skey []byte

func Init(a config.Authorization) error {
	if a.Skey == "" {
		return errors.New("empty skey")
	}
	skey = []byte(a.Skey)
	return nil
}

func GetSkey() []byte { return skey }

func GenerateToken(name string) (string, int64, error) {
	tokenExpirationDuration := config.Configs.Constants.TokenExpirationDuration
	expireAt := time.Now().Add(tokenExpirationDuration).Unix()
	c := &Claims{
		Username: name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireAt,
			Issuer:    name,
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	s, err := t.SignedString(skey)
	if err != nil {
		return "", 0, err
	}
	return s, expireAt, nil
}

// claims, err := auth.GetClaimsFromHeader(c)
func GetClaimsFromHeader(c *response.ResponseContext) (Claims, error) {
	tokenHeaderName := config.Configs.Constants.TokenHeaderName
	bearer := strings.Split(c.Request().Header.Get(tokenHeaderName), " ")
	if len(bearer) < 2 {
		return Claims{}, errors.New("invalid header")
	}
	if bearer[0] != "Bearer" {
		return Claims{}, errors.New("invalid header")
	}

	raw := bearer[1]

	if model.CheckDeprecatedToken(SuffixExpiredToken+raw) == true { // this token is invalid
		return Claims{}, errors.New("invalid token")
	}

	claims := Claims{}
	_, err := jwt.ParseWithClaims(raw, &claims,
		func(token *jwt.Token) (interface{}, error) {
			return GetSkey(), nil
		})
	if err != nil {
		return Claims{}, err
	}
	claims.RawToken = raw
	return claims, nil
}

const SuffixEmailVerifyCode string = "emailVerifyCode:"

// err = auth.LoginEmailSend(name, email)
func LoginEmailSend(name string, email string) error {
	conf := config.Configs.SmtpConfig
	sender, password := conf.Sender, conf.Password
	host, port := conf.Host, conf.Port
	m := gomail.NewMessage()
	m.SetHeader("From", sender)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "BBYD Verification Code")

	var code string
	var err error
	for i := 1; i <= 10; i++ { // retry for 10 times
		code, err = generator.GenerateEmailVerifyCode()
		key := SuffixEmailVerifyCode + code
		if err != nil {
			return err
		}

		err = model.SetEmailVerifyCode(key, name, conf.CodeExpirationMinute*60)
		if err == nil {
			break
		}
	}
	if code == "" {
		return errors.New("cannot generate verification code")
	}

	m.SetBody("text/html",
		"You are trying to login user <"+name+"> by email verification.</br>"+
			"Here is your code:</br>"+
			`<font color="blueviolet" size="18px">`+code+"</font></br>"+
			"The code is valid for "+strconv.Itoa(conf.CodeExpirationMinute)+" minutes.")
	d := gomail.NewDialer(host, port, sender, password)

	err = d.DialAndSend(m)
	if err != nil {
		return err
	}
	return nil
}

// name, err := auth.CodeNameGet(code)
func CodeNameGet(code string) (string, error) {
	res, err := model.CodeNameGet(SuffixEmailVerifyCode + code)
	if err != nil {
		return "", model.RedisInternalError
	}
	if res == nil {
		return "", model.RedisNotFoundError
	}
	return string(res.([]uint8)), nil
}

const SuffixExpiredToken string = "expiredToken:"

// err := auth.LogoutToken(rawToken)
func LogoutToken(rawToken, name string) error {
	return model.SetExpiredToken(SuffixExpiredToken+rawToken, name)
}
