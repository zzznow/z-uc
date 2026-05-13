package internal

import (
	"fmt"
	LOGGER "log/slog"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var Db *sqlx.DB

func InitMysql(cfg *MysqlConfig) (err error) {
	fmt.Println(cfg.Host + ":" + strconv.Itoa(cfg.Port))
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True",
		cfg.User,
		cfg.Passwd,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)
	Db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		LOGGER.Error("connect mysql failed", "error", err.Error())
		return err
	}
	Db.SetMaxOpenConns(cfg.MaxOpenConns)
	Db.SetMaxIdleConns(cfg.MaxIdleConns)

	err = Db.Ping()
	if err != nil {
		LOGGER.Error("ping mysql failed", "error", err.Error())
	}

	return
}

func CloseMysql() {
	_ = Db.Close()
}
