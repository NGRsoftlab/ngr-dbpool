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
				fmt.Sprintf(
					"%s://%s:%s@%s/%s",
					"testD1",
					"testI",
					"testP",
					"testU",
					"testPS",
				): {
					DB:     &sql.DB{},
					Mapper: nil,
				},
				fmt.Sprintf(
					"%s://%s:%s@%s/%s",
					"testD2",
					"testI",
					"testP",
					"testU",
					"testPS",
				): {
					DB:     &sql.DB{},
					Mapper: nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
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

			},
		)
	}
}

func TestSafeDbMapCache(t *testing.T) {
	cacheDuration := 500 * time.Millisecond
	cleanupInterval := 100 * time.Millisecond

	cache := New(cacheDuration, cleanupInterval)
	defer cache.ClearAll()

	tests := []struct {
		name      string
		key       string
		value     *sqlx.DB
		duration  time.Duration
		sleepTime time.Duration
		exists    bool
	}{
		{
			name:     "set and get valid entry",
			key:      "db_1",
			value:    &sqlx.DB{},
			duration: 10 * time.Second,
			exists:   true,
		},
		{
			name:      "check expired entry",
			key:       "expired_db",
			value:     &sqlx.DB{},
			duration:  1 * time.Second,
			sleepTime: 2 * time.Second,
			exists:    false,
		},
		{
			name:     "zero duration (use default expiration)",
			key:      "default_duration",
			value:    &sqlx.DB{},
			duration: 0,
			exists:   true,
		},
		{
			name:   "non-existent entry",
			key:    "no_db",
			exists: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if tt.value != nil {
					cache.Set(tt.key, tt.value, tt.duration)
				}

				if tt.sleepTime > 0 {
					time.Sleep(tt.sleepTime)
				}

				got, ok := cache.Get(tt.key)
				if ok != tt.exists {
					t.Errorf("Get() got=%t, want=%t", ok, tt.exists)
				}

				if ok && got != tt.value {
					t.Errorf("Get() value missmatch, got=%v, want=%v", got, tt.value)
				}
			},
		)
	}
}

func TestClearAll(t *testing.T) {
	cache := New(5*time.Second, 1*time.Second)
	cache.Set("db1", &sqlx.DB{}, 5*time.Second)
	cache.Set("db2", &sqlx.DB{}, 5*time.Second)

	if len(cache.GetItems()) != 2 {
		t.Errorf("GetItems() length expected to be 2")
	}

	cache.ClearAll()

	if len(cache.GetItems()) != 0 {
		t.Errorf("cache.ClearAll() failed, items count %d, want 0", len(cache.GetItems()))
	}
}

func TestDeleteEntry(t *testing.T) {
	cache := New(5*time.Second, 1*time.Second)
	dbKey := "db_to_delete"
	cache.Set(dbKey, &sqlx.DB{}, 5*time.Second)

	err := cache.Delete(dbKey)
	if err != nil {
		t.Errorf("unexpected error on Delete: %v", err)
	}

	_, ok := cache.Get(dbKey)
	if ok {
		t.Errorf("Expected entry to be deleted")
	}

	err = cache.Delete("non_existent_key")
	if err == nil {
		t.Errorf("Expected error when deleting non-existent key")
	}
}
