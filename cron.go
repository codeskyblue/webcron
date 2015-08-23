package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/go-xorm/xorm"
	"github.com/robfig/cron"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

var xe *xorm.Engine

const (
	TRIGGER_MANUAL   = "manual"
	TRIGGER_SCHEDULE = "schedule"
)

type Task struct {
	Name        string            `json:"name"`
	Schedule    string            `json:"schedule"`
	Command     string            `json:"command"`
	Description string            `json:"description"`
	Environ     map[string]string `json:"environ"`
	Enabled     bool              `json:"enabled"`
}

func (task *Task) Run(trigger string) (err error) {
	_, rec, err := keeper.NewRecord(task.Name)
	if err != nil {
		return err
	}
	rec.Trigger = trigger
	switch runtime.GOOS {
	case "windows":
		err = execute(rec, "cmd", []string{"/c", task.Command})
	case "linux":
		fallthrough
	default:
		err = execute(rec, "/bin/bash", []string{"-c", task.Command})
	}
	return
}

func execute(rec *Record, command string, args []string) (err error) {
	start := time.Now()
	defer func() {
		rec.Duration = time.Since(start)
		keeper.DoneRecord(rec.Key())
	}()
	//log.Printf("executing: %s %s", command, strings.Join(args, " "))

	cmd := exec.Command(command, args...)
	cmd.Stdout = io.MultiWriter(os.Stdout, rec.Buffer)
	cmd.Stderr = io.MultiWriter(os.Stderr, rec.Buffer)
	for k, v := range rec.T.Environ {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	if err = cmd.Start(); err != nil {
		rec.ExitCode = 130
		return err
	}
	// extrace exit_code from err
	if err = cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				log.Printf("Exit Status: %d", status.ExitStatus())
				rec.ExitCode = status.ExitStatus()
				return err
			}
		}
		rec.ExitCode = 131
	}
	return nil
}

func create() (cr *cron.Cron, wgr *sync.WaitGroup) {
	/*var schedule string = os.Args[1]
	var command string = os.Args[2]
	var args []string = os.Args[3:len(os.Args)]

	*/
	wg := &sync.WaitGroup{}

	c := cron.New()
	//println("new cron:", schedule)

	return c, wg
}

func loadTasks(filename string) ([]Task, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var tasks []Task
	err = json.Unmarshal(data, &tasks)
	return tasks, err
}

type Record struct {
	Id        int64
	Name      string `xorm:"unique(nt)"`
	Index     int    `xorm:"unique(nt)"`
	Trigger   string
	ExitCode  int
	CreatedAt time.Time `xorm:"created"`
	Duration  time.Duration
	T         Task `xorm:"-"` //`xorm:"task"` //FIXME(ssx): when use task will not get xorm work

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
