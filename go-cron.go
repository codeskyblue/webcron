package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/robfig/cron"
)

func execute(command string, args []string) {
	log.Printf("executing: %s %s", command, strings.Join(args, " "))

	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func create() (cr *cron.Cron, wgr *sync.WaitGroup) {
	var schedule string = os.Args[1]
	var command string = os.Args[2]
	var args []string = os.Args[3:len(os.Args)]

	wg := &sync.WaitGroup{}

	c := cron.New()
	println("new cron:", schedule)

	c.AddFunc(schedule, func() {
		wg.Add(1)
		execute(command, args)
		wg.Done()
	})

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

func main() {
	flag.Parse()

	c, wg := create()
	go start(c, wg)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	stop(c, wg)
}
