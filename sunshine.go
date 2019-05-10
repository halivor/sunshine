package main

import (
	"log"
	"sync"

	evp "github.com/halivor/goevent/eventpool"
	mw "github.com/halivor/goevent/middleware"
	ac "github.com/halivor/sunshine/acceptor"
	ag "github.com/halivor/sunshine/agent"
	_ "github.com/halivor/sunshine/transfer"
)

var wg sync.WaitGroup

func main() {
	wg.Add(1)
	go newSun("0.0.0.0:10301", "127.0.0.1:10205")
	go newSun("0.0.0.0:10302", "127.0.0.1:10205")
	go newSun("0.0.0.0:10303", "127.0.0.1:10205")
	wg.Wait()
}

func newSun(laddr, raddr string) {
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
	if _, e := ac.NewTcpAcceptor(laddr, ep, mws); e != nil {
		log.Println("new acceptor failed:", e)
	}
	// TODO:  multi agent/eventpool cause buffer pool crash
	if _, e := ag.New(raddr, ep, mws); e != nil {
		log.Println("new agent failed:", e)
	}
	ep.Run()
}
