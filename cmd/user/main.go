package main

import (
	"user/service"
)

func main() {
	service.New(
		service.Option{
			ListenPort:":8080",
		}).Run()
}