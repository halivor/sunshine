package main

import (
	evp "frontend/eventpool"
	ls "frontend/listener"
)

func main() {
	ep := evp.New()
	ls.NewTcpListener("0.0.0.0:19981", ep)
	ep.Run()
}
