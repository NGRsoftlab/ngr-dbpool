package dbpool

import (
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

// GetConnectionByParams - get *sqlx.DB from cache (if exists) or create new and put into cache
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
		return nil, err
	}

	//db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(duration)
	db.Mapper = reflectx.NewMapperFunc("json", func(s string) string { return s })

	//set conn to connCache
	connCache.Set(connString, db, duration)

	conn, ok = connCache.Get(connString)
	if !ok && conn == nil {
		return nil, errors.New("no conn in connCache")
	}

	return conn, nil
}
