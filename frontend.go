package main

import (
	"log"
	"sync"

	ac "github.com/halivor/frontend/acceptor"
	ag "github.com/halivor/frontend/agent"
	_ "github.com/halivor/frontend/transfer"
	evp "github.com/halivor/goevent/eventpool"
	mw "github.com/halivor/goevent/middleware"
)

var wg sync.WaitGroup

func main() {
	wg.Add(1)
	go newFrontend("0.0.0.0:10301")
	go newFrontend("0.0.0.0:10302")
	go newFrontend("0.0.0.0:10303")
	wg.Wait()
}

func newFrontend(addr string) {
	defer func() {
		/*if r := recover(); r != nil {*/
		//log.Println("panic =>", r)
		/*}*/
		wg.Done()
	}()
	ep, e := evp.New()
	if e != nil {
		log.Println("new event pool failed:", e)
	}
	mws := mw.New()
	if _, e := ac.NewTcpAcceptor(addr, ep, mws); e != nil {
		log.Println("new acceptor failed:", e)
	}
	if _, e := ag.New("127.0.0.1:10205", ep, mws); e != nil {
		log.Println("new agent failed:", e)
	}
	ep.Run()
}
