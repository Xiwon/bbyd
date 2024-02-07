package auth

import (
	// "net/smtp"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
	// "strconv"
	mathRand "math/rand"

	"bbyd/internal/shared/config"
	"bbyd/pkg/utils/response"

	"github.com/dgrijalva/jwt-go"
	"github.com/gophish/gomail"
)

type Claims struct {
	Username string `json:"username"`
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
	claims := Claims{}
	_, err := jwt.ParseWithClaims(raw, &claims,
		func(token *jwt.Token) (interface{}, error) {
			return GetSkey(), nil
		})
	if err != nil {
		return Claims{}, err
	}
	return claims, nil
}

// salt := auth.GetSaltFromSecret(db_sec)
func GetSaltFromSecret(s string) string { return s[:strings.Index(s, "$")] }

// sec := auth.GenerateSecret(req.Passwd, salt)
func GenerateSecret(passwd string, salt string) string {
	sha := fmt.Sprintf("%x", sha256.Sum256([]byte(passwd)))
	return salt + "$" +
		fmt.Sprintf("%x", md5.Sum([]byte(sha+salt)))
}

// salt := auth.GenerateSalt()
func GenerateSalt() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// [deprecated]
func GenerateRand6() string {
	mathRand.NewSource(time.Now().Unix())
	return fmt.Sprintf("%6d", mathRand.Intn(1000000))
}

// err = auth.LoginEmailSend(name, email)
func LoginEmailSend(name string, email string) error {
	conf := config.Configs.SmtpConfig
	sender, password := conf.Sender, conf.Password
	host, port := conf.Host, conf.Port
	m := gomail.NewMessage()
	m.SetHeader("From", sender)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "BBYD Email Login")
	code, _, err := GenerateToken(name)
	if err != nil {
		return err
	}
	m.SetBody("text/html", "Please visit route /user/token/email/"+code+" to login your account")
	d := gomail.NewDialer(host, port, sender, password)

	err = d.DialAndSend(m)
	if err != nil {
		return err
	}
	return nil
}