package config

import (
	"github.com/BurntSushi/toml"
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
	PreRegisterRoot bool   `toml:"preregisterroot"`
}
type Authorization struct {
	Skey string `toml:"skey"`
}

type Config struct {
	Server        Server        `toml:"server"`
	Database      Database      `toml:"database"`
	Authorization Authorization `toml:"authorization"`
}

func Create() error {
	_, err := toml.DecodeFile("../config/config.toml", Configs)
	return err
}
