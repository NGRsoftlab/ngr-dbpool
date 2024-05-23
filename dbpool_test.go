package dbpool

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

func TestDbPoolCache(t *testing.T) {
	tests := []struct {
		name         string
		okResTimeout time.Duration
		noResTimeout time.Duration
		checkData    map[string]*sqlx.DB
	}{
		{
			name:         "valid (nil dn.connections in data)",
			okResTimeout: 1 * time.Second,
			noResTimeout: 6 * time.Second,
			checkData: map[string]*sqlx.DB{
				fmt.Sprintf("%s://%s:%s@%s/%s",
					"testD1",
					"testI",
					"testP",
					"testU",
					"testPS"): {
					DB:     &sql.DB{},
					Mapper: nil,
				},
				fmt.Sprintf("%s://%s:%s@%s/%s",
					"testD2",
					"testI",
					"testP",
					"testU",
					"testPS"): {
					DB:     &sql.DB{},
					Mapper: nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LocalCache := New(10*time.Second, 5*time.Second)
			defer LocalCache.ClearAll()

			for k, v := range tt.checkData {
				LocalCache.Set(k, v, 5*time.Second)
			}

			time.Sleep(tt.okResTimeout)
			for k := range tt.checkData {
				_, ok := LocalCache.Get(k)
				if !ok {
					t.Error("no cached result found")
				}
			}

			time.Sleep(tt.noResTimeout)
			for k := range tt.checkData {
				_, ok := LocalCache.Get(k)
				if ok {
					t.Error("cached result was not deleted after timeout")
				}
			}

		})
	}
}
