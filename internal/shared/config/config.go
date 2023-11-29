package config

import (
	"github.com/BurntSushi/toml"
	"time"
)

var Configs *Config = &Config{}

type Server struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
}
type Database struct {
	Host            string `toml:"host"`
	Port            int    `toml:"port"`
	User            string `toml:"user"`
	Password        string `toml:"password"`
	DbName          string `toml:"dbname"`
	SslMode         bool   `toml:"sslmode"`
	PreRegisterRoot bool   `toml:"preregister_root"`
}
type Authorization struct {
	Skey string `toml:"skey"`
}

type Constants struct {
	TokenHeaderName         string `toml:"token_header_name"`
	TokenExpirationMinute   int    `toml:"token_expiration_minute"`
	TokenExpirationDuration time.Duration

	ApiPathRoot string `toml:"api_path_root"`

	AdminAuthname     string `toml:"admin_authname"`
	DefaultAuthname   string `toml:"default_authname"`
	RootName          string `toml:"root_name"`
	RootDefaultPasswd string `toml:"root_default_password"`
}

type Config struct {
	Server        Server        `toml:"server"`
	Database      Database      `toml:"database"`
	Authorization Authorization `toml:"authorization"`
	Constants     Constants     `toml:"constants"`
}

func Create() error {
	_, err := toml.DecodeFile("../config/config.toml", Configs)
	if err != nil {
		return err
	}

	Configs.Constants.TokenExpirationDuration =
		time.Duration(Configs.Constants.TokenExpirationMinute) * time.Minute
	return nil
}
