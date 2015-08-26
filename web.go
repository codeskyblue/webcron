// main
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/qiniu/log"

	"github.com/Unknwon/macaron"
	"github.com/go-xorm/xorm"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

var m = macaron.Classic()
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	Subprotocols:    []string{"cron"},
}

func init() {
	m.Use(macaron.Renderer(macaron.RenderOptions{
		Delims: macaron.Delims{"[[", "]]"},
	}))
}

func struct2JS(v interface{}) template.JS {
	data, _ := json.Marshal(v)
	return template.JS(string(data))
}

func initRoutes() {
	m.Get("/", func(ctx *macaron.Context) {
		// data, _ := json.Marshal(keeper.Tasks())
		// ctx.Data["Tasks"] = template.JS(string(data))
		// rds, _ := keeper.ListRecords(8)
		rds, _ := keeper.ListUniqRecords()
		// data, _ := json.Marshal(rds)
		if len(rds) > 0 {
			latest := rds[0]
			ctx.Data["Current"] = struct2JS(latest)
		}
		ctx.Data["Records"] = struct2JS(rds)
		ctx.HTML(200, "homepage")
	})

	m.Get("/:name", func(ctx *macaron.Context) {
		// keeper.GetRecord(name, index)
		name := ctx.Params(":name")
		// rds, _ := keeper.ListUniqRecords()
		rds, _ := keeper.ListRecords(name, 8)
		latest := rds[0]
		ctx.Data["Current"] = struct2JS(latest)
		ctx.Data["Records"] = struct2JS(rds)
		ctx.HTML(200, "homepage")
	})

	m.Get("/settings", func(ctx *macaron.Context) {
		ctx.Data["Fresh"] = template.JS("true")
		ctx.HTML(200, "settings")
	})

	m.Get("/:name/builds/:index", func(ctx *macaron.Context) {
		name := ctx.Params(":name")
		index := ctx.ParamsInt(":index")
		rec, _ := keeper.GetRecord(name, index)
		rds, _ := keeper.ListRecords(name, 8)
		ctx.Data["Current"] = struct2JS(rec)
		ctx.Data["Records"] = struct2JS(rds)
		ctx.HTML(200, "homepage")
	})

	m.Get("/settings/:name", func(ctx *macaron.Context) {
		name := ctx.Params(":name")
		task, created := keeper.GetOrCreateTask(name)
		ctx.Data["Task"] = struct2JS(task)
		ctx.Data["Fresh"] = struct2JS(created)
		ctx.HTML(200, "settings")
	})

	m.Get("/api/records/:name/index/:index", func(ctx *macaron.Context) {
		name := ctx.Params(":name")
		index := ctx.ParamsInt(":index")
		rec, err := keeper.GetRecord(name, index)
		if err != nil {
			ctx.Error(500, err.Error())
		}
		ctx.JSON(200, rec)
	})

	m.Get("/api/tasks", func(ctx *macaron.Context) {
		rds, err := keeper.ListUniqRecords()
		if err != nil {
			ctx.Error(500, err.Error())
		}
		ctx.JSON(200, rds)
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
		//log.Println(task)
		// FIXME(ssx): need change
		task.Enabled = true
		if err := keeper.PutTask(task.Name, task); err != nil {
			ctx.Error(500, err.Error())
			return
		}
		ctx.JSON(200, "Task modified")
	})

	m.Get("/ws/:name/:index", func(w http.ResponseWriter, r *http.Request, ctx *macaron.Context) {
		name := ctx.Params(":name")
		index := ctx.ParamsInt(":index")
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer ws.Close()

		wsend := func(tp string, body []byte) error {
			return ws.WriteJSON(map[string]string{
				"type": tp,
				"body": string(body),
			})
		}

		rec, err := keeper.GetRecord(name, index)
		if err != nil {
			log.Println(err)
			ws.Close()
			return
		}
		if !rec.Running {
			body, err := rec.LogData()
			if err != nil {
				wsend("error", []byte(err.Error()))
				return
			}
			wsend("whole", body)
			return
		}
		currBody, rd := rec.wb.NewBufReader(r.RemoteAddr)
		defer rd.Close()
		wsend("before", currBody)
		var buf = make([]byte, 1000)
		for {
			cnt, err := rd.Read(buf)
			if err != nil && err == io.EOF {
				wsend("stream", buf[:cnt])
				wsend("finish", []byte(fmt.Sprintf("%d", rec.ExitCode)))
				break
			}
			if err != nil {
				wsend("error", []byte(err.Error()))
				break
			}
			wsend("stream", buf[:cnt])
		}
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
	xe, err = xorm.NewEngine("sqlite3", "./test.db")
	//xe, err = xorm.NewEngine("mysql", "cron:cron@tcp(10.246.13.180:3306)/cron?charset=utf8")
	// xe, err = xorm.NewEngine("mysql", "root:@/cron?charset=utf8")
	if err != nil {
		log.Fatal(err)
	}
	if err := xe.Sync(Record{}); err != nil {
		log.Fatal(err)
	}

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
	http.Handle("/", m)
	//http.Handle("/-/", http.StripPrefix("/-/", http.FileServer(http.Dir("public"))))
	if err := http.ListenAndServe(":"+strconv.Itoa(gcfg.ServerPort), nil); err != nil {
		log.Fatal(err)
	}
}
