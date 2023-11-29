package model

import(
	"strconv"
	"bbyd/internal/shared/config"
	"bbyd/pkg/utils/logs"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"go.uber.org/zap"
)

func Init(d config.Database) error {
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

	return AutoMigrate(d)
}