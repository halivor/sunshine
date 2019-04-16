package eventpool

import (
	"fmt"
	"log"
	"syscall"

	cnf "github.com/halivor/frontend/config"
)

type EventPool interface {
	AddEvent(ev Event) error
	ModEvent(ev Event) error
	DelEvent(ev Event) error
	Run()
	Stop()
}

type eventpool struct {
	fd   int
	ev   []syscall.EpollEvent // 每次被唤醒，最大处理event数
	ss   map[int]Event        // pool中的event
	stop bool
	*log.Logger
}

func init() {
}

func New() (EventPool, error) {
	fd, e := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	switch e {
	case nil:
		return &eventpool{
			fd:     fd,
			ev:     make([]syscall.EpollEvent, cnf.MaxEvents),
			ss:     make(map[int]Event, cnf.MaxConns),
			stop:   false,
			Logger: cnf.NewLogger(fmt.Sprintf("[ep(%d)] ", fd)),
		}, nil
	default:
		// EINVAL (epoll_create1()) Invalid value specified in flags.
		// EMFILE The per-user limit on the number of epoll instances imposed by
		//        /proc/sys/fs/epoll/max_user_instances was encountered.  See
		//        epoll(7) for further details.
		// EMFILE The per-process limit on the number of open file descriptors
		//        has been reached.
		// ENFILE The system-wide limit on the total number of open files has
		//        been reached.
		// ENOMEM There was insufficient memory to create the kernel object.
		return nil, e
	}
}

func (m *eventpool) AddEvent(ev Event) error {
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
	case syscall.EBADF, syscall.EEXIST, syscall.EINVAL,
		syscall.ENOMEM, syscall.ENOSPC, syscall.EPERM:
		return e
	default:
		// 该epoll已不可用需要重建
		return e
	}
}

func (m *eventpool) ModEvent(ev Event) error {
	m.Println("mod event", ev.Fd(), ev.Event())
	switch e := syscall.EpollCtl(m.fd,
		syscall.EPOLL_CTL_MOD,
		ev.Fd(),
		&syscall.EpollEvent{
			Events: ev.Event(),
			Fd:     int32(ev.Fd()),
		},
	); e {
	case syscall.EBADF, syscall.EINVAL, syscall.ENOENT,
		syscall.ENOMEM, syscall.EPERM:
		return e
	default:
		// 该epoll已不可用需要重建
		return e
	}
}

func (m *eventpool) DelEvent(ev Event) error {
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
	case syscall.EBADF, syscall.EINVAL, syscall.ENOENT,
		syscall.EEXIST, syscall.ENOSPC, syscall.EPERM:
		return e
	default:
		// 该epoll已不可用需要重建
		return e
	}

}

func (m *eventpool) Run() {
	for !m.stop {
		switch n, e := syscall.EpollWait(m.fd, m.ev, 1000); e {
		case syscall.EINTR:
		case nil:
			for i := 0; i < n; i++ {
				//m.Println(event[es[m.ev[i].Events]])
				m.ss[int(m.ev[i].Fd)].CallBack(m.ev[i].Events)
			}
		default:
			// 该epoll已不可用需要重建
			// EBADF  epfd is not a valid file descriptor.
			// EFAULT The memory area pointed to by events is not accessible with
			//        write permissions.
			// EINVAL epfd is not an epoll file descriptor, or maxevents is less
			//        than or equal to zero.
		}
	}
	m.Println("event pool stop")
}

func (m *eventpool) Stop() {
	m.stop = true
}
