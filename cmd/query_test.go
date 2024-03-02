package main

import (
	"database/sql"
	"testing"

	"github.com/Buff2out/sqlite-go-one/config/log"
)

func BenchmarkQuery(b *testing.B) {
	sugar, logger := log.GetSugaredLogger()
	defer logger.Sync()
	db, err := sql.Open("sqlite", "src/newvideo.db")
	if err != nil {
		sugar.Fatalw("Error to connect DB", "MSG", err)
	}
	defer db.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.QueryRow("SELECT title, views FROM videos WHERE video_id = ?", "R39-E3uG5J0")
	}
}
