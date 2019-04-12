package main

import (
	"log"
	"sync"

	ac "github.com/halivor/frontend/acceptor"
	ag "github.com/halivor/frontend/agent"
	evp "github.com/halivor/frontend/eventpool"
	m "github.com/halivor/frontend/middleware"
)

var wg sync.WaitGroup

func main() {
	wg.Add(1)
	go newFrontEnd()
	wg.Wait()
}

func newFrontEnd() {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
		wg.Done()
	}()
	ep, e := evp.New()
	if e != nil {
		return
	}
	mw := m.New()
	ac.NewTcpAcceptor("0.0.0.0:19981", ep, mw)
	ag.New("127.0.0.1:9981", ep, mw)
	ep.Run()
	wg.Done()
}
