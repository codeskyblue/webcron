package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	exit := flag.Int("e", 0, "exit code")
	flag.Parse()
	for i := 0; i < 10; i++ {
		fmt.Println("Now:", i)
		time.Sleep(time.Second)
	}
	os.Exit(*exit)
}
