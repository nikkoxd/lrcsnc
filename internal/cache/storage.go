package cache

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"lrcsnc/internal/pkg/global"
	"lrcsnc/internal/pkg/log"
	playerStructs "lrcsnc/internal/pkg/structs/player"
	"lrcsnc/internal/pkg/types"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"

	_ "github.com/mattn/go-sqlite3"
)

const (
	_DB_NAME string = "lrcsnc.db"
)

type Storage struct {
	db *sql.DB
}

var StorageInstance Storage

func Init() {
	var err error
	StorageInstance.db, err = sql.Open("sqlite3", os.ExpandEnv(global.Config.C.Cache.Dir)+"/"+_DB_NAME)
	if err != nil {
		log.Fatal("cache", fmt.Sprintf("Failed to open SQLite database at %v; can be an issue with permissions or a faulty database. More: %v", os.ExpandEnv(global.Config.C.Cache.Dir)+"/"+_DB_NAME, err))
		return
	}
	err = StorageInstance.db.Ping()
	if err != nil {
		log.Fatal("cache", fmt.Sprintf("Failed to ping SQLite database at %v; can be a faulty database. More: %v", os.ExpandEnv(global.Config.C.Cache.Dir)+"/"+_DB_NAME, err))
		return
	}

	q, args := createTableQuery()
	_, err = StorageInstance.db.Exec(q, args...)
	if err != nil {
		log.Fatal("cache", fmt.Sprintf("Failed to ensure the availability of lrcsnc_cache table; can be a faulty database. More: %v", err))
		return
	}
	q, args = createIndexQuery()
	_, err = StorageInstance.db.Exec(q, args...)
	if err != nil {
		log.Fatal("cache", fmt.Sprintf("Failed to ensure the availability of lrcsnc_cache_idx index; can be a faulty database. More: %v", err))
		return
	}
}

func Close() {
	_ = StorageInstance.db.Close()
}

func (s *Storage) Fetch(song *playerStructs.Song) (playerStructs.LyricsData, CacheState) {
	global.Config.M.Lock()
	defer global.Config.M.Unlock()

	var lyrData string
	var lyrState byte
	var updatedAt time.Time

	q, args := selectQuery(song)
	err := s.db.QueryRow(q, args...).Scan(&lyrData, &lyrState, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		log.Debug("cache", "Failed to fetch cached lyrics; no such entry is present (cache miss).")
		return playerStructs.LyricsData{LyricsState: types.LyricsStateNotFound}, CacheStateNonExistant
	} else if err != nil {
		log.Error("cache", fmt.Sprintf("Failed to fetch cached lyrics; error: %v", err))
		return playerStructs.LyricsData{LyricsState: types.LyricsStateNotFound}, CacheStateNonExistant
	}

	var lyrics playerStructs.Lyrics
	err = json.Unmarshal([]byte(lyrData), &lyrics)
	if err != nil {
		log.Error("cache", fmt.Sprintf("Failed to unmarshal lyrics data from cache; error: %v", err))
		return playerStructs.LyricsData{LyricsState: types.LyricsStateNotFound}, CacheStateNonExistant
	}

	var cacheState CacheState = CacheStateExpired
	if global.Config.C.Cache.LifeSpan == 0 || time.Since(updatedAt).Hours() < float64(global.Config.C.Cache.LifeSpan) {
		cacheState = CacheStateActive
	}

	return playerStructs.LyricsData{Lyrics: lyrics, LyricsState: types.LyricsState(lyrState)}, cacheState
}

func (s *Storage) Store(song *playerStructs.Song) error {
	q, args := insertQuery(song)
	_, err := s.db.Exec(q, args...)
	if err != nil {
		log.Error("cache", fmt.Sprintf("Failed to store cached lyrics; error: %v", err))
	}

	return err
}

func (s *Storage) Remove(song *playerStructs.Song) error {
	q, args := removeQuery(song)
	_, err := s.db.Exec(q, args...)
	if err != nil {
		log.Error("cache", fmt.Sprintf("Failed to remove cached lyrics; error: %v", err))
	}

	return err
}
