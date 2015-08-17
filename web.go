package main

import "github.com/Unknwon/macaron"

var m = macaron.Classic()

func init() {
	m.Use(macaron.Renderer())
}

func initRoutes() {
	m.Get("/", func(ctx *macaron.Context) {
		ctx.HTML(200, "index")
	})
}

func main() {
	initRoutes()
	m.Run()
}
