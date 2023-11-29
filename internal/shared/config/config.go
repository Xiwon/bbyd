package config

import (
	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"
	"time"
)

var Configs *Config = &Config{}
var validate *validator.Validate = validator.New()

type Server struct {
	Host string `toml:"host"`
	Port int    `toml:"port" validate:"required,gte=1024,lt=65536"`
}
type Database struct {
	Host            string `toml:"host"`
	Port            int    `toml:"port"     validate:"required,gte=0,lt=65536"`
	User            string `toml:"user"     validate:"required"`
	Password        string `toml:"password"`
	DbName          string `toml:"dbname"   validate:"required"`
	SslMode         bool   `toml:"sslmode"`
	PreRegisterRoot bool   `toml:"preregister_root"`
}
type Authorization struct {
	Skey string `toml:"skey" validate:"required"`
}

type Constants struct {
	TokenHeaderName         string `toml:"token_header_name"`
	TokenExpirationMinute   int    `toml:"token_expiration_minute" validate:"required"`
	TokenExpirationDuration time.Duration

	ApiPathRoot string `toml:"api_path_root" validate:"required"`

	AdminAuthname     string `toml:"admin_authname"        validate:"required"`
	DefaultAuthname   string `toml:"default_authname"      validate:"required"`
	RootName          string `toml:"root_name"             validate:"required"`
	RootDefaultPasswd string `toml:"root_default_password" validate:"required"`
}

type Config struct {
	Server        Server        `toml:"server"`
	Database      Database      `toml:"database"`
	Authorization Authorization `toml:"authorization"`
	Constants     Constants     `toml:"constants"`
}

func Create() error {
	// default settings
	Configs.Server.Host = "localhost"

	Configs.Database.Host = "localhost"
	Configs.Database.SslMode = false
	Configs.Database.PreRegisterRoot = true

	Configs.Constants.TokenHeaderName = "Authorization"

	_, err := toml.DecodeFile("../config/config.toml", Configs)
	if err != nil {
		return err
	}
	err = validate.Struct(*Configs)
	if err != nil {
		return err
	}

	// after work
	Configs.Constants.TokenExpirationDuration =
		time.Duration(Configs.Constants.TokenExpirationMinute) * time.Minute

	return nil
}
