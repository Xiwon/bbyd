package config

type Server struct {
	Host string
	Port int
}
type Database struct {
	Host string
	Port int
	User string
	Password string
	DbName string
	SslMode bool 
}
type Authorization struct {
	Skey string
}

type Config struct {
	Se Server
	Db Database
	Au Authorization
}

func Create() (c Config, e error) {
	c.Se = Server{
		"localhost",
		11451,
	}
	c.Db = Database{
		"localhost",
		5432,
		"postgres",
		"",
		"byddb",
		false,
	}
	c.Au = Authorization{
		"bbingyan_jwt_skey_58490998",
	}
	
	return
}