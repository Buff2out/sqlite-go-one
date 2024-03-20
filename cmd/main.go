package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Buff2out/sqlite-go-one/config/log"
	"github.com/jmoiron/sqlx"

	_ "modernc.org/sqlite"
)

type Video struct {
	Id          string
	Title       string
	PublishTime time.Time
	Tags        []string
	Views       int
}

type TagVideo struct {
	Tags string
}

func GetList(ctx context.Context, db *sqlx.DB) (videos []TagVideo, err error) {
	sqlSelect := `SELECT tags FROM videos 
                    WHERE tags LIKE '%worst%' GROUP BY tags`
	err = db.SelectContext(ctx, &videos, sqlSelect)
	return
}

func retrieveVideoCSVToDB(ctx context.Context, db *sql.DB, csvFile string) error {
	file, errOpen := os.Open(csvFile)
	if errOpen != nil {
		return errOpen
	}
	defer file.Close()
	var videos []Video = make([]Video, 0, 1000)

	const (
		Id          = 0
		Title       = 2
		PublishTime = 5
		Tags        = 6
		Views       = 7
	)
	r := csv.NewReader(file)
	// пропускаем первую строку с неймингом полей
	if _, errReadline := r.Read(); errReadline != nil {
		return errReadline
	}
	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		v := Video{
			Id:    line[Id],
			Title: line[Title],
		}
		if v.PublishTime, err = time.Parse(time.RFC3339, line[PublishTime]); err != nil {
			return err
		}
		tags := strings.Split(line[Tags], "|")
		for i, val := range tags {
			tags[i] = strings.Trim(val, `"`)
			/*
				в учебнике почему то торчит v. Понять не могу.
					tags[i] = strings.Trim(v, `"`)

				теперь понял, это значение массива. Но лучше уж не перезаписывать v Video
			*/
		}
		v.Tags = tags
		if v.Views, err = strconv.Atoi(line[Views]); err != nil {
			return err
		}
		videos = append(videos, v)
		if len(videos) == 1000 {
			if err = insertVideos(ctx, db, videos); err != nil {
				return err
			}
			videos = videos[:0]
		}
	}
	return insertVideos(ctx, db, videos)
}

func insertVideos(ctx context.Context, db *sql.DB, videos []Video) error {
	/*
		пока что без возвращаемого значения LastInsertId() и RowsAffected()
	*/
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, errPrep := tx.PrepareContext(ctx,
		"INSERT INTO videos (video_id, title, publish_time, tags, views) "+
			"VALUES ($1, $2, $3, $4, $5)")

	if errPrep != nil {
		return errPrep
	}
	defer stmt.Close()

	for _, val := range videos {
		_, errExec := stmt.ExecContext(ctx, val.Id, val.Title, val.PublishTime, strings.Join(val.Tags, "|"), val.Views)
		if errExec != nil {
			return err
		}
	}

	return tx.Commit()
}

func SQLCreateTableVideos(db *sql.DB) error {
	q := `CREATE TABLE IF NOT EXISTS videos (
		"id" INTEGER PRIMARY KEY AUTOINCREMENT,
		"video_id" TEXT,
		"title" TEXT,
		"publish_time" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		"tags" TEXT,
		"views" INTEGER NOT NULL DEFAULT 0
	)`
	_, err := db.Exec(q)
	return err
}

func main() {
	sugar, logger := log.GetSugaredLogger()
	defer logger.Sync()
	db := sqlx.MustOpen("sqlite", "newvideo.db")
	defer db.Close()

	ctx := context.Background()
	start := time.Now()
	videos, err := GetList(ctx, db)
	if err != nil {
		sugar.Fatalw("err in GetList", "err", err)
		return
	}

	sqlUpdate := "UPDATE videos SET tags = ? WHERE tags = ?"

	var updates int64
	for _, val := range videos {
		var newTags []string
		for _, tag := range strings.Split(val.Tags, "|") {
			if !strings.Contains(strings.ToLower(tag), "best") {
				newTags = append(newTags, tag)
			}
		}
		res := db.MustExecContext(ctx, sqlUpdate, strings.Join(newTags, `|`), val.Tags)
		// посмотрим, сколько записей было обновлено
		if upd, err := res.RowsAffected(); err == nil {
			updates += upd
		}
	}

	sugar.Infow("SUCCESS", "Затраченное время", time.Since(start))
	sugar.Infow("total operations", "updates", updates)
}
