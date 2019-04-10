package eventpool

import (
	"fmt"
	"log"
	"os"
	"syscall"

	"frontend/config"
)

type EventPool interface {
	AddEvent(ev Event) error
	ModEvent(ev Event) error
	DelEvent(ev Event) error
}

type eventpoll struct {
	fd int
	ev []syscall.EpollEvent
	ss map[int]Event
	*log.Logger
}

func init() {
}

func New() *eventpoll {
	fd, err := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	if err != nil {
		log.Panicln(err.Error())
	}
	return &eventpoll{
		fd:     fd,
		ev:     make([]syscall.EpollEvent, config.MaxEvents),
		ss:     make(map[int]Event, config.MaxConns),
		Logger: log.New(os.Stderr, fmt.Sprintf("[ep(%d)] ", fd), log.LstdFlags|log.Lmicroseconds),
	}
}

func (m *eventpoll) AddEvent(ev Event) error {
	m.ss[ev.Fd()] = ev
	m.Println("add event", ev.Fd(), ev.Event())
	return syscall.EpollCtl(m.fd,
		syscall.EPOLL_CTL_ADD,
		ev.Fd(),
		&syscall.EpollEvent{
			Events: ev.Event(),
			Fd:     int32(ev.Fd()),
		},
	)
}

func (m *eventpoll) ModEvent(ev Event) error {
	m.Println("mod event", ev.Fd(), ev.Event())
	return syscall.EpollCtl(m.fd,
		syscall.EPOLL_CTL_MOD,
		ev.Fd(),
		&syscall.EpollEvent{
			Events: ev.Event(),
			Fd:     int32(ev.Fd()),
		},
	)
}

func (m *eventpoll) DelEvent(ev Event) error {
	m.Println("del event", ev.Fd(), ev.Event())
	delete(m.ss, ev.Fd())
	return syscall.EpollCtl(m.fd,
		syscall.EPOLL_CTL_DEL,
		ev.Fd(),
		&syscall.EpollEvent{
			Events: ev.Event(),
			Fd:     int32(ev.Fd()),
		},
	)
}

func (m *eventpoll) Run() {
	for {
		n, err := syscall.EpollWait(m.fd, m.ev, 0)
		if err != nil {
			log.Panicln(err.Error())
		}
		for i := 0; i < n; i++ {
			m.Println(event[m.ev[i].Events])
			m.ss[int(m.ev[i].Fd)].CallBack(m.ev[i].Events)
		}
	}
}
