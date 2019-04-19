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
}

type eventpool struct {
	fd   int
	ev   []syscall.EpollEvent // 每次被唤醒，最大处理event数
	es   map[int]Event        // pool中的event
	stop bool
	*log.Logger
}

func New() (*eventpool, error) {
	return newEp(nil)
}

func newEp(epo *eventpool) (*eventpool, error) {
	fd, e := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	switch e {
	case nil:
		if epo == nil {
			epo = &eventpool{
				fd:     fd,
				ev:     make([]syscall.EpollEvent, cnf.MaxEvents),
				es:     make(map[int]Event, cnf.MaxConns),
				stop:   false,
				Logger: cnf.NewLogger(fmt.Sprintf("[ep(%d)] ", fd)),
			}
		} else {
			epo.fd = fd
			epo.Logger = cnf.NewLogger(fmt.Sprintf("[ep(%d)] ", fd))
		}
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
	return epo, nil
}

func (ep *eventpool) AddEvent(ev Event) error {
	ep.es[ev.Fd()] = ev
	ep.Println("add event", ev.Fd(), ev.Event())
	switch e := syscall.EpollCtl(ep.fd,
		syscall.EPOLL_CTL_ADD,
		ev.Fd(),
		&syscall.EpollEvent{
			Events: ev.Event(),
			Fd:     int32(ev.Fd()),
		},
	); e {
	case syscall.EBADF, syscall.EEXIST, syscall.EINVAL, syscall.ELOOP,
		syscall.ENOMEM, syscall.ENOSPC, syscall.EPERM:
		return e
	default:
		// 不应该有其他错误信息
		ep.Println(ev, e)
		return e
	}
}

func (ep *eventpool) ModEvent(ev Event) error {
	ep.Println("mod event", ev.Fd(), ev.Event())
	switch e := syscall.EpollCtl(ep.fd,
		syscall.EPOLL_CTL_MOD,
		ev.Fd(),
		&syscall.EpollEvent{
			Events: ev.Event(),
			Fd:     int32(ev.Fd()),
		},
	); e {
	case syscall.EBADF, syscall.EINVAL, syscall.ELOOP, syscall.ENOENT,
		syscall.ENOSPC, syscall.ENOMEM, syscall.EPERM:
		return e
	default:
		// 理论上不存在其他错误信息
		ep.Println(ev, e)
		return e
	}
}

func (ep *eventpool) DelEvent(ev Event) error {
	ep.Println("del event", ev.Fd(), ev.Event())
	delete(ep.es, ev.Fd())
	switch e := syscall.EpollCtl(ep.fd,
		syscall.EPOLL_CTL_DEL,
		ev.Fd(),
		&syscall.EpollEvent{
			Events: ev.Event(),
			Fd:     int32(ev.Fd()),
		},
	); e {
	case syscall.EBADF, syscall.EINVAL, syscall.ELOOP, syscall.ENOENT,
		syscall.ENOSPC, syscall.ENOMEM, syscall.EPERM:
		return e
	default:
		// 理论上不存在其他错误信息
		ep.Println(ev, e)
		return e
	}

}

func (ep *eventpool) Run() {
	for !ep.stop {
		switch n, e := syscall.EpollWait(ep.fd, ep.ev, cnf.EP_TIMEOUT); e {
		case syscall.EINTR:
		case nil:
			for i := 0; i < n; i++ {
				//ep.Println(event[es[ep.ev[i].Events]])
				ep.es[int(ep.ev[i].Fd)].CallBack(ep.ev[i].Events)
			}
		default:
			// 该epoll已不可用需要重建
			// EBADF  epfd is not a valid file descriptor.
			// EFAULT The memory area pointed to by events is not accessible with
			//        write permissions.
			// EINVAL epfd is not an epoll file descriptor, or maxevents is less
			//        than or equal to zero.

			ep.Println("epoll wait error", e)
			// 理论上不存在，若存在则直接重建
			if e := ep.reNew(); e != nil {
				ep.Println("ep run failed,", e)
				return
			}
		}
	}
	ep.Println("event pool stop")
}

func (ep *eventpool) Release() {
	for _, ev := range ep.es {
		ev.Release()
	}
}

// TODO: 考虑细节
func (ep *eventpool) reNew() error {
	syscall.Close(ep.fd)
	if _, e := newEp(ep); e != nil {
		return e
	}
	for _, ev := range ep.es {
		if e := ep.AddEvent(ev); e != nil {
			ev.Release()
		}
	}
	return nil
}

func (ep *eventpool) Stop() {
	ep.stop = true
}
