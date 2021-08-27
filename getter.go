package dbpool

import (
	errorCustom "github.com/NGRsoftlab/error-lib"

	"context"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"time"
)

func GetConnectionByParams(Ctx context.Context, connCache *SafeDbMapCache,
	duration time.Duration, driver, connString string) (*sqlx.DB, error) {

	conn, ok := connCache.Get(connString)
	if ok && conn != nil {
		// ping to check
		err := conn.PingContext(Ctx)
		if err != nil {
			return nil, err
		}

		return conn, nil
	}

	//create conn
	db, err := sqlx.ConnectContext(Ctx, driver, connString)
	if err != nil {
		return nil, errorCustom.GlobalErrors.ErrBadDbConn()
	}

	//db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(duration)
	db.Mapper = reflectx.NewMapperFunc("json", func(s string) string { return s })

	//set conn to connCache
	connCache.Set(connString, db, duration)

	conn, ok = connCache.Get(connString)
	if !ok && conn == nil {
		return nil, errorCustom.GlobalErrors.ErrBadDbConn()
	}

	return conn, nil
}
