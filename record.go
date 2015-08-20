package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var db gorm.DB

func init() {
	var err error
	db, err = gorm.Open("mysql", "root:@/cron")
	if err != nil {
		log.Fatal(err)
	}
	if err = db.DB().Ping(); err != nil {
		log.Fatal(err)
	}
	db.CreateTable(&Record{})
}

const (
	TRIGGER_MANUAL   = "manual"
	TRIGGER_SCHEDULE = "schedule"
)

type Record struct {
	ID        int
	Name      string
	Trigger   string
	Index     int
	ExitCode  int
	CreatedAt time.Time
	Duration  time.Duration

	Buffer  *bytes.Buffer `sql:"-"`
	Running bool          `sql:"-"`
}

func (r *Record) Key() string {
	return fmt.Sprintf("%s:%d", r.Name, r.Index)
}

func (r *Record) LogPath() string {
	if r.Index == -1 {
		return filepath.Join("logs", r.Name+"-latest.log")
	}
	return filepath.Join("logs", fmt.Sprintf("%s-%d.log", r.Name, r.Index))
}

func (r *Record) Done() (err error) {
	err = ioutil.WriteFile(r.LogPath(), r.Buffer.Bytes(), 0644)
	if err != nil {
		return
	}
	r.Running = false
	db.Create(r)
	// FIXME(ssx): save to db
	return nil
}
