// +build bindata

package main

import (
	"log"
	"net/http"
)

func init() {
	log.Println("Enable bindata fs")
	http.Handle("/-/", http.StripPrefix("/-/", http.FileServer(assetFS())))
}
