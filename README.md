# ngr-dbpool in-memory db-connection-cache implementation (based on sqlx db connections inside)

# import
```
import "github.com/NGRsoftlab/ngr-dbpool"
```

# example
```
// create db conn cache with default 10sec storing and 5sec garbage collecting interval
LocalDBCache := New(10*time.Second, 5*time.Second)
defer LocalCache.ClearAll()

connTimeOut := 5*time.Second
execTimeOut := 55*time.Second

Ctx, cancel := context.WithTimeout(ctx, execTimeOut)
defer cancel()

db, err := dbpool.GetConnectionByParams(Ctx, LocalDBCache,
	connTimeOut,
	"postgres", "<postgres connection string>")
if err != nil {
	return err
}

err = db.SelectContext(Ctx, result, db.Rebind("<sql query to get some information>"), args...)
if err != nil {
	return err
}
// ... some other logic
```