package main

import (
	"sync"

	ep "github.com/halivor/goutility/eventpool"
	log "github.com/halivor/goutility/logger"
	mw "github.com/halivor/goutility/middleware"
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
		//log.Warn("panic =>", r)
		/*}*/
		wg.Done()
	}()
	log.StdOutDebug()
	eper := ep.New()
	mws := mw.New()
	if _, e := ac.NewTcpAcceptor(laddr, eper, mws); e != nil {
		log.Warn("new acceptor failed:", e)
	}
	// TODO:  multi agent/eventpool cause buffer pool crash
	if _, e := ag.New(raddr, eper, mws); e != nil {
		log.Warn("new agent failed:", e)
	}
	eper.Run()
}
