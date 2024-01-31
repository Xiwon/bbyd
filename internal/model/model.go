package model

import(
	"errors"
	"strconv"
	"bbyd/internal/shared/config"
	"bbyd/pkg/utils/logs"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"go.uber.org/zap"
	"github.com/garyburd/redigo/redis"
)

func Init(d config.Database, r config.RedisConfig) error {
	if d.Port == r.Port {
		return errors.New("database port crashed")
	}

	// Postgresql database init
	dsn := 
		" host=" + d.Host +
		" port=" + strconv.Itoa(d.Port) +
		" user=" + d.User +
		" password=" + d.Password +
		" dbname=" + d.DbName
	if d.SslMode {
		dsn += " sslmode=enable"
	} else {
		dsn += " sslmode=disable"
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logs.Error("database connection failed. ", zap.Error(err))
		return err
	}

	err = AutoMigrate(d)
	if err != nil {
		return err
	}

	// redis init
	redisConn, err = redis.Dial("tcp", r.Host+":"+strconv.Itoa(r.Port))
	if err != nil {
		return err
	}
	return nil
}