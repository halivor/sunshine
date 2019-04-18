package main

import (
	"log"
	"sync"

	ac "github.com/halivor/frontend/acceptor"
	ag "github.com/halivor/frontend/agent"
	evp "github.com/halivor/frontend/eventpool"
	mw "github.com/halivor/frontend/middleware"
	_ "github.com/halivor/frontend/middleware/transfer"
)

var wg sync.WaitGroup

func main() {
	wg.Add(1)
	go newFrontend()
	wg.Wait()
}

// 20190418 -> 单核CPU，35000连接，每秒4条消息，CPU达到100%
// 20190419 -> 单核CPU， 8000连接，每秒20条消息，CPU达到100%
func newFrontend() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("panic =>", r)
		}
		wg.Done()
	}()
	ep, e := evp.New()
	if e != nil {
		log.Println("new event pool failed:", e)
	}
	mws := mw.New()
	if _, e := ac.NewTcpAcceptor("0.0.0.0:10000", ep, mws); e != nil {
		log.Println("new acceptor failed:", e)
	}
	if _, e := ag.New("127.0.0.1:29981", ep, mws); e != nil {
		log.Println("new agent failed:", e)
	}
	ep.Run()
}
