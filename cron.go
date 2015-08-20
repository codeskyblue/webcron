package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/robfig/cron"
)

type Task struct {
	Name        string `json:"name"`
	Schedule    string `json:"schedule"`
	Command     string `json:"command"`
	Description string `json:"description"`
}

func (task *Task) Run() (err error) {
	switch runtime.GOOS {
	case "windows":
		err = execute("cmd", []string{"/c", task.Command})
	case "linux":
		fallthrough
	default:
		err = execute("/bin/bash", []string{"-c", task.Command})
	}
	return
}

func execute(command string, args []string) error {
	return nil
	log.Printf("executing: %s %s", command, strings.Join(args, " "))

	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func create() (cr *cron.Cron, wgr *sync.WaitGroup) {
	/*var schedule string = os.Args[1]
	var command string = os.Args[2]
	var args []string = os.Args[3:len(os.Args)]

	*/
	wg := &sync.WaitGroup{}

	c := cron.New()
	//println("new cron:", schedule)

	for _, task := range tasks {
		ta := task // make a copy, this is necessary
		taskFunc := func() {
			wg.Add(1)
			defer wg.Done()
			if err := ta.Run(); err != nil {
				log.Println(ta.Name, err)
			}
		}
		c.AddFunc(task.Schedule, taskFunc)
	}

	return c, wg
}

func start(c *cron.Cron, wg *sync.WaitGroup) {
	c.Start()
}

func stop(c *cron.Cron, wg *sync.WaitGroup) {
	println("Stopping")
	c.Stop()
	println("Waiting")
	wg.Wait()
	println("Exiting")
	os.Exit(0)
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

/*
func main() {
	flag.Parse()

	c, wg := create()
	go start(c, wg)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	stop(c, wg)
}
*/
