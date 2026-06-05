package mysql

import (
	"context"
	"fmt"
	"time"

	drivermysql "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	appconfig "github.com/ww1489/WarSpark/internal/config"
)

func New(ctx context.Context, cfg appconfig.MySQLConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open("mysql", DSN(cfg))
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpen)
	db.SetMaxIdleConns(cfg.MaxIdle)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func DSN(cfg appconfig.MySQLConfig) string {
	return buildMySQLDSN(cfg)
}

func MigrationURL(cfg appconfig.MySQLConfig) string {
	return "mysql://" + buildMySQLDSN(cfg)
}

func WithinTx(ctx context.Context, db *sqlx.DB, fn func(*sqlx.Tx) error) (err error) {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			_ = tx.Rollback()
			panic(recovered)
		}
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}
	err = tx.Commit()
	return err
}

func buildMySQLDSN(cfg appconfig.MySQLConfig) string {
	loc := time.Local
	if cfg.Loc != "" && cfg.Loc != "Local" {
		if parsed, err := time.LoadLocation(cfg.Loc); err == nil {
			loc = parsed
		}
	}

	dsn := drivermysql.NewConfig()
	dsn.User = cfg.User
	dsn.Passwd = cfg.Password
	dsn.Net = "tcp"
	dsn.Addr = fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	dsn.DBName = cfg.DB
	dsn.ParseTime = cfg.ParseTime
	dsn.Loc = loc
	dsn.Params = map[string]string{
		"charset": cfg.Charset,
	}

	return dsn.FormatDSN()
}
