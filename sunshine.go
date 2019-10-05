package main

import (
	"sync"

	ep "github.com/halivor/goutility/eventpool"
	mw "github.com/halivor/goutility/middleware"
	ac "github.com/halivor/sunshine/acceptor"
	ag "github.com/halivor/sunshine/agent"
	up "github.com/halivor/sunshine/peer"
)

var wg sync.WaitGroup

func main() {
	wg.Add(3)
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
	mws := mw.New()
	eper := ep.New()
	mgr := up.NewManager(mws)
	mgr.BindMsg(0)
	ac.NewTcpAcceptor(laddr, eper, mgr)
	ag.New(raddr, eper, mws)
	eper.Run()
}
