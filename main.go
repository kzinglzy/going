package main

import (
	"fmt"
	"going"
)

func main() {
	server, err := going.NewServer("192.168.1.1:9191")
	if err != nil {
		fmt.Println(err)
	}
	defer server.Close()
}
