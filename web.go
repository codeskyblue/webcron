package main

import (
	"encoding/json"
	"flag"
	"html/template"
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
		data, _ := json.Marshal(tasks)
		ctx.Data["Tasks"] = template.JS(string(data))
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
		log.Println(task)
		ctx.JSON(200, "New task has been added")
	})
}

var (
	cfgFile = flag.String("c", "config.json", "crontab setting file")
	srvPort = flag.Int("p", 4000, "port to listen")
	logDir  = flag.String("logdir", "logs", "log directory")
	tasks   []Task
)

func main() {
	flag.Parse()

	var err error
	//xe, err = xorm.NewEngine("sqlite3", "./test.db")
	xe, err = xorm.NewEngine("mysql", "root:@/cron?charset=utf8")
	if err != nil {
		log.Fatal(err)
	}
	xe.Sync(Record{})

	if _, err = os.Stat(*logDir); err != nil {
		os.Mkdir(*logDir, 0755)
	}
	tasks, err = loadTasks(*cfgFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(tasks)

	keeper = NewKeeper(tasks)

	key, rec, err := keeper.NewRecord("test1")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(key, rec)

	if err := keeper.DoneRecord(key); err != nil {
		log.Fatal(err)
	}

	cr, _ := create()
	cr.Start()

	initRoutes()
	log.Printf("Listening on *:%d", *srvPort)
	http.ListenAndServe(":"+strconv.Itoa(*srvPort), m)
}
