package main
import "os"
import "os/exec"
import "strings"
import "sync"
import "os/signal"
import "syscall"
import "github.com/robfig/cron"

func execute(command string, args []string)(output string, e error) {

    println("executing:", command, strings.Join(args, " "))

    cmd := exec.Command(command, args...)
    out, err := cmd.Output()

    if err != nil {
        return "", err
    }

    return string(out), nil
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
        out, err := execute(command, args)
        wg.Done()
        if err != nil {
            println(err.Error())
        }

        println(out)
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

    c, wg := create()

    go start(c, wg)

    ch := make(chan os.Signal, 1)
    signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
    println(<-ch)

    stop(c, wg)
}


