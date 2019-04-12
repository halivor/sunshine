package eventpool

import (
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/halivor/frontend/config"
)

type EventPool interface {
	AddEvent(ev Eventer) error
	ModEvent(ev Eventer) error
	DelEvent(ev Eventer) error
}

type eventpoll struct {
	fd int
	ev []syscall.EpollEvent
	ss map[int]Eventer
	*log.Logger
}

func init() {
}

func New() (*eventpoll, error) {
	fd, e := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	switch e {
	case nil:
		return &eventpoll{
			fd:     fd,
			ev:     make([]syscall.EpollEvent, config.MaxEvents),
			ss:     make(map[int]Eventer, config.MaxConns),
			Logger: log.New(os.Stderr, fmt.Sprintf("[ep(%d)] ", fd), log.LstdFlags|log.Lmicroseconds),
		}, nil
	default:
		return nil, e
	}
}

func (m *eventpoll) AddEvent(ev Eventer) error {
	m.ss[ev.Fd()] = ev
	m.Println("add event", ev.Fd(), ev.Event())
	switch e := syscall.EpollCtl(m.fd,
		syscall.EPOLL_CTL_ADD,
		ev.Fd(),
		&syscall.EpollEvent{
			Events: ev.Event(),
			Fd:     int32(ev.Fd()),
		},
	); e {
	case syscall.EEXIST, syscall.ENOMEM, syscall.ENOSPC, syscall.EPERM:
		return e
	default:
		// 该epoll已不可用需要重建
		return e
	}
}

func (m *eventpoll) ModEvent(ev Eventer) error {
	m.Println("mod event", ev.Fd(), ev.Event())
	switch e := syscall.EpollCtl(m.fd,
		syscall.EPOLL_CTL_MOD,
		ev.Fd(),
		&syscall.EpollEvent{
			Events: ev.Event(),
			Fd:     int32(ev.Fd()),
		},
	); e {
	case syscall.ENOENT, syscall.EEXIST, syscall.ENOMEM, syscall.ENOSPC, syscall.EPERM:
		return e
	default:
		// 该epoll已不可用需要重建
		return e
	}
}

func (m *eventpoll) DelEvent(ev Eventer) error {
	m.Println("del event", ev.Fd(), ev.Event())
	delete(m.ss, ev.Fd())
	switch e := syscall.EpollCtl(m.fd,
		syscall.EPOLL_CTL_DEL,
		ev.Fd(),
		&syscall.EpollEvent{
			Events: ev.Event(),
			Fd:     int32(ev.Fd()),
		},
	); e {
	case syscall.ENOENT, syscall.EEXIST, syscall.ENOMEM, syscall.ENOSPC, syscall.EPERM:
		return e
	default:
		// 该epoll已不可用需要重建
		return e
	}

}

func (m *eventpoll) Run() {
	for {
		switch n, e := syscall.EpollWait(m.fd, m.ev, 0); e {
		case syscall.EINTR:
		case nil:
			for i := 0; i < n; i++ {
				m.Println(event[es[m.ev[i].Events]])
				m.ss[int(m.ev[i].Fd)].CallBack(m.ev[i].Events)
			}
		default:
			// 该epoll已不可用需要重建
		}
	}
}
