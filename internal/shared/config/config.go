package config

type Server struct {
	Host string
	Port int
}
type Database struct {
	Host            string
	Port            int
	User            string
	Password        string
	DbName          string
	SslMode         bool
	PreRegisterRoot bool
}
type Authorization struct {
	Skey string
}

type Config struct {
	Server        Server
	Database      Database
	Authorization Authorization
}

func Create() (c Config, e error) {
	c.Server = Server{
		Host: "localhost",
		Port: 11451,
	}
	c.Database = Database{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "",
		DbName:   "byddb",
		SslMode:  false,

		PreRegisterRoot: true,
	}
	c.Authorization = Authorization{
		Skey: "bbingyan_jwt_skey_58490998",
	}

	return
}
