package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/go-xorm/xorm"
)

var xe *xorm.Engine

const (
	TRIGGER_MANUAL   = "manual"
	TRIGGER_SCHEDULE = "schedule"
)

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("%d", time.Time(t).Unix())
	return []byte(stamp), nil
}

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
		rec.wb.CloseWriters()
		rec.Duration = time.Since(start)
		keeper.DoneRecord(rec.Key())
	}()
	//log.Printf("executing: %s %s", command, strings.Join(args, " "))

	rec.wb = NewWriteBroadcaster()
	rec.Running = true

	cmd := exec.Command(command, args...)
	cmd.Stdout = rec.wb
	cmd.Stderr = rec.wb
	// cmd.Stdout = io.MultiWriter(os.Stdout, rec.wb)
	// cmd.Stderr = io.MultiWriter(os.Stderr, rec.wb)
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
				// log.Printf("Exit Status: %d", status.ExitStatus())
				rec.ExitCode = status.ExitStatus()
				return err
			}
		}
		rec.ExitCode = 131
	}
	return nil
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
	Name      string        `json:"name" xorm:"unique(nt)"`
	Index     int           `json:"index" xorm:"unique(nt)"`
	Trigger   string        `json:"trigger"`
	ExitCode  int           `json:"exit_code"`
	CreatedAt JSONTime      `json:"created_at" xorm:"created"`
	Duration  time.Duration `json:"duration"`
	T         Task          `json:"task" xorm:"json task"`

	Buffer  *bytes.Buffer     `json:"-" xorm:"-"`
	Running bool              `json:"running" xorm:"-"`
	wb      *WriteBroadcaster `json:"-" xorm:"-"`
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

func (r *Record) LogData() ([]byte, error) {
	return ioutil.ReadFile(r.LogPath())
}

func (r *Record) Done() (err error) {
	err = ioutil.WriteFile(r.LogPath(), r.wb.Bytes(), 0644)
	if err != nil {
		return
	}
	r.Running = false
	_, err = xe.InsertOne(r)
	return
}
