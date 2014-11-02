package main

import (
	"fmt"
	http2 "github.com/zoncoen/minimum_http2"
	"os"
)

func main() {
	args := os.Args
	if len(args) != 3 {
		fmt.Println("Usage: [DEBUG=1] go run filename {ADDR} {PORT}")
		os.Exit(1)
	}
	client := &http2.Client{}
	err := client.Connect(args[1] + ":" + args[2])
	if err != nil {
		panic(err.Error())
	}

	res := client.Get("/")
	fmt.Println(string(res.Body))

	client.Disconnect()
}
