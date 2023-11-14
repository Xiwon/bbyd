package auth

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"bbyd/internal/shared/config"
	"bbyd/pkg/utils/response"

	"github.com/dgrijalva/jwt-go"
)

const (
	tokenHeaderName         = "Authorization"
	tokenExpirationDuration = 30 * time.Minute
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

func getSkey() []byte { return skey }

func GenerateToken(name string) (string, int64, error) {
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
			return getSkey(), nil
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
