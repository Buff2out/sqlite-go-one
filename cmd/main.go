package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Buff2out/sqlite-go-one/config/log"

	_ "modernc.org/sqlite"
)

type Video struct {
	Id          string
	Title       string
	PublishTime time.Time
	Tags        []string
	Views       int
}

func readVideoCSV(csvFile string) ([]Video, error) {
	file, errOpen := os.Open(csvFile)
	if errOpen != nil {
		return nil, errOpen
	}
	defer file.Close()
	var videos []Video

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
		return nil, errReadline
	}
	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		v := Video{
			Id:    line[Id],
			Title: line[Title],
		}
		if v.PublishTime, err = time.Parse(time.RFC3339, line[PublishTime]); err != nil {
			return nil, err
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
			return nil, err
		}
		videos = append(videos, v)
	}
	return videos, nil
}

func insertVideos(ctx context.Context, db *sql.DB, videos []Video) error {
	/*
		Загатова
	*/
	return nil
}

func main() {
	sugar, logger := log.GetSugaredLogger()
	defer logger.Sync()
	db, err := sql.Open("sqlite", "src/video.db")
	if err != nil {
		sugar.Infow("CONN ERR", "KeyErr", err)
	}
	defer db.Close()

	videos, err := readVideoCSV("src/USvideos.csv")
	if err != nil {
		sugar.Fatalw("Error to readVideoCSV() ", "errMsg", err)
	}
	err = insertVideos(context.Background(), db, videos)
	if err != nil {
		sugar.Fatalw("Error to insertVideos() ", "errMsg", err)
	}
	sugar.Infow(fmt.Sprintf("Всего csv-записей %v\n", len(videos)))
}
