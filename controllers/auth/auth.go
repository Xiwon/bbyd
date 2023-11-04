package auth

import(
	"fmt"
	"time"
	"errors"
	"strings"
	"crypto/sha256"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
		
	"bbyd/shared/config"

	"github.com/dgrijalva/jwt-go"
)

type TokenClaims struct {
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

func GenerateToken(name string) string {
	c := TokenClaims{
		Username: name,
		StandardClaims: jwt.StandardClaims{
        	NotBefore: time.Now().Unix() - 60,
        	ExpiresAt: time.Now().Unix() + 60,
        	Issuer: name,
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	s, _ := t.SignedString(skey)
	return string(s)
}

// salt := auth.GetSaltFromSecret(db_sec)
func GetSaltFromSecret(s string) string { return s[ : strings.Index(s, "$")] }

// sec := auth.GenerateSecret(req.Passwd, salt)
func GenerateSecret(passwd string, salt string) string {
	sha := fmt.Sprintf("%x", sha256.Sum256([]byte(passwd)))
	return salt + "$" + 
		fmt.Sprintf("%x", md5.Sum([]byte(sha + salt)))
}

// salt := auth.GenerateSalt()
func GenerateSalt() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}