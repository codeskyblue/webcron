package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/Unknwon/macaron"
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
}

var (
	cfgFile = flag.String("c", "config.json", "crontab setting file")
	srvPort = flag.Int("p", 4000, "port to listen")
	tasks   []Task
)

func main() {
	flag.Parse()

	var err error
	tasks, err = loadTasks(*cfgFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(tasks)
	cr, _ := create()
	cr.Start()

	initRoutes()
	log.Printf("Listening on *:%d", *srvPort)
	http.ListenAndServe(":"+strconv.Itoa(*srvPort), m)
}
