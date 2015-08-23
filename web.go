package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/qiniu/log"

	"github.com/Unknwon/macaron"
	"github.com/go-xorm/xorm"
)

var m = macaron.Classic()

func init() {
	m.Use(macaron.Renderer(macaron.RenderOptions{
		Delims: macaron.Delims{"[[", "]]"},
	}))
}

func initRoutes() {
	m.Get("/", func(ctx *macaron.Context) {
		//data, _ := json.Marshal(tasks)
		//ctx.Data["Tasks"] = template.JS(string(data))
		ctx.HTML(200, "index")
	})

	m.Get("/settings", func(ctx *macaron.Context) {
		ctx.HTML(200, "settings")
	})

	m.Post("/api/tasks", func(ctx *macaron.Context) {
		var task Task
		dec := json.NewDecoder(ctx.Req.Body().ReadCloser())
		if err := dec.Decode(&task); err != nil {
			ctx.Error(500, err.Error())
			return
		}
		if err := keeper.AddTask(task); err != nil {
			ctx.Error(500, err.Error())
			return
		}
		log.Println(task)
		ctx.JSON(200, "New task has been added")
	})

	m.Put("/api/tasks", func(ctx *macaron.Context) {
		var task Task
		dec := json.NewDecoder(ctx.Req.Body().ReadCloser())
		if err := dec.Decode(&task); err != nil {
			ctx.Error(500, err.Error())
			return
		}
		// FIXME(ssx): need change
		task.Enabled = true
		if err := keeper.PutTask(task.Name, task); err != nil {
			ctx.Error(500, err.Error())
			return
		}
		ctx.JSON(200, "Task modified")
	})
}

type GlobalConfig struct {
	SchedFile  string
	LogDir     string
	ServerPort int
}

var (
	gcfg GlobalConfig
)

func main() {
	flag.StringVar(&gcfg.SchedFile, "sched", "sched.json", "file which store schedule setting")
	flag.IntVar(&gcfg.ServerPort, "port", 4000, "port to listen")
	flag.StringVar(&gcfg.LogDir, "logdir", "logs", "log directory")
	flag.Parse()

	var err error
	//xe, err = xorm.NewEngine("sqlite3", "./test.db")
	xe, err = xorm.NewEngine("mysql", "root:@/cron?charset=utf8")
	if err != nil {
		log.Fatal(err)
	}
	//xe.Sync(Record{})

	if _, err = os.Stat(gcfg.LogDir); err != nil {
		os.Mkdir(gcfg.LogDir, 0755)
	}
	if _, err = os.Stat(gcfg.SchedFile); err != nil {
		ioutil.WriteFile(gcfg.SchedFile, []byte("[]"), 0644)
	}
	tasks, err := loadTasks(gcfg.SchedFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(tasks)

	keeper = NewKeeper(tasks)

	initRoutes()
	log.Printf("Listening on *:%d", gcfg.ServerPort)
	http.ListenAndServe(":"+strconv.Itoa(gcfg.ServerPort), m)
}
