package dbpool

import (
	_ "github.com/lib/pq"
	_ "github.com/mailru/go-clickhouse"

	. "github.com/NGRsoftlab/ngr-logging"

	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"testing"
	"time"
)

var okTimeout = time.Second * 60

func TestDbPoolCache(t *testing.T) {
	LocalCache := New(30*time.Second, 5*time.Second)

	defer LocalCache.ClearAll()

	////CH
	//driver := "clickhouse"
	//connStr := fmt.Sprintf("http://%s:%s@%s:%s/%s?read_timeout=20s&write_timeout=30s",
	//	"",
	//	"",
	//	"",
	//	"8123",
	//	"")
	//
	//Ctx, cancel := context.WithTimeout(context.Background(), okTimeout)
	//defer cancel()
	//
	//db, err := sqlx.ConnectContext(Ctx, driver, connStr)
	//if err != nil {
	//	logging.Logger.Fatal(err)
	//}
	//
	//LocalCache.Set(connStr, db, 10*time.Second)
	//
	//time.Sleep(5 * time.Second)
	//
	//cachedRes, ok := LocalCache.Get(connStr)
	//if ok {
	//	logging.Logger.Debug("cached db is here: ", cachedRes)
	//
	//	// use cachedRes
	//	var name string
	//	err = cachedRes.GetContext(Ctx, &name,"SELECT name FROM test WHERE id=1")
	//	if err != nil {
	//		logging.Logger.Fatal(err)
	//	}
	//
	//} else {
	//	logging.Logger.Debug("no res: ", connStr)
	//}
	//
	//time.Sleep(6 * time.Second)
	//
	//cachedRes, ok = LocalCache.Get(connStr)
	//if ok {
	//	logging.Logger.Debug("cached db is here: ", cachedRes)
	//
	//	// use cachedRes
	//	var name string
	//	err = cachedRes.GetContext(Ctx, &name,"SELECT name FROM test WHERE id=1")
	//	if err != nil {
	//		logging.Logger.Fatal(err)
	//	}
	//
	//} else {
	//	logging.Logger.Debug("no res: ", connStr)
	//}

	////////////////////////////////////////////////////////////////////////

	//// POSTGRES
	driver := "postgres"
	connStr := fmt.Sprintf("%s://%s:%s@%s/%s",
		"postgres",
		"", // username
		"", // userpass
		"", // ip
		"") // dbname

	Ctx, cancel := context.WithTimeout(context.Background(), okTimeout)
	defer cancel()

	db, err := sqlx.ConnectContext(Ctx, driver, connStr)
	if err != nil {
		Logger.Fatal(err)
	}

	LocalCache.Set(connStr, db, 10*time.Second)

	//time.Sleep(5 * time.Second)

	cachedRes, ok := LocalCache.Get(connStr)
	if ok {
		Logger.Debug("cached db is here: ", cachedRes)

		// use cachedRes
		var name string
		err = cachedRes.GetContext(Ctx, &name, "SELECT name FROM test WHERE id=1")
		if err != nil {
			Logger.Fatal(err)
		}

	} else {
		Logger.Debug("no res: ", connStr)
	}

	time.Sleep(5 * time.Second)

	cachedRes, ok = LocalCache.Get(connStr)
	if ok {
		Logger.Debug("cached db is here: ", cachedRes)

		// use cachedRes
		var name string
		err = cachedRes.GetContext(Ctx, &name, "SELECT name FROM test WHERE id=1")
		if err != nil {
			Logger.Fatal(err)
		}

	} else {
		Logger.Debug("no res: ", connStr)
	}
}
