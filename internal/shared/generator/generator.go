package generator

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"

	"bbyd/internal/shared/config"
)

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

const CharSet string = `!"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_abcdefghijklmnopqrstuvwxyz{|}~`

var CharSetSize big.Int = *big.NewInt(int64(len(CharSet)))

// code, err := generator.GenerateEmailVerifyCode()
func GenerateEmailVerifyCode() (string, error) {
	len := config.Configs.SmtpConfig.CodeLength

	var randMax, digit big.Int
	randMax.Exp(&CharSetSize, big.NewInt(int64(len)), nil) // base**len integer
	res, err := rand.Int(rand.Reader, &randMax)
	if err != nil {
		return "", err
	}

	b := make([]byte, len)
	for i := 0; i < len; i++ {
		res.DivMod(res, &CharSetSize, &digit)
		b[i] = CharSet[digit.Int64()]
	}

	return string(b), nil
}
