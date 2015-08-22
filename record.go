package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	//_ "github.com/mattn/go-sqlite3"
)

var xe *xorm.Engine

const (
	TRIGGER_MANUAL   = "manual"
	TRIGGER_SCHEDULE = "schedule"
)

type Record struct {
	Id        int64
	Name      string `xorm:"unique(nt)"`
	Index     int    `xorm:"unique(nt)"`
	Trigger   string
	ExitCode  int
	CreatedAt time.Time `xorm:"created"`
	Duration  time.Duration
	T         Task `xorm:"'task'"`

	Buffer  *bytes.Buffer `xorm:"-"`
	Running bool          `xorm:"-"`
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
	_, err = xe.InsertOne(r)
	return
}
