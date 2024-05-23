package dbpool

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	. "github.com/NGRsoftlab/ngr-logging"

	"github.com/jmoiron/sqlx"
)

func TestDbPoolCache(t *testing.T) {
	LocalCache := New(30*time.Second, 5*time.Second)
	defer LocalCache.ClearAll()

	connStr1 := fmt.Sprintf("%s://%s:%s@%s/%s",
		"testD1",
		"testI",
		"testP",
		"testU",
		"testPS")
	connStr2 := fmt.Sprintf("%s://%s:%s@%s/%s",
		"testD2",
		"testI",
		"testP",
		"testU",
		"testPS")

	db := &sqlx.DB{
		DB:     &sql.DB{},
		Mapper: nil,
	}

	LocalCache.Set(connStr1, db, 10*time.Second)
	LocalCache.Set(connStr2, db, 10*time.Second)

	cachedRes, ok := LocalCache.Get(connStr1)
	if ok {
		Logger.Debugf("cached db is here: %+v", cachedRes)
	} else {
		Logger.Debugf("no res: %s", connStr1)
	}

	//time.Sleep(5 * time.Second)
	//
	//cachedRes, ok = LocalCache.Get(connStr)
	//if ok {
	//	Logger.Debugf("cached db is here: %+v", cachedRes)
	//} else {
	//	Logger.Debugf("no res: %s", connStr)
	//}
}
