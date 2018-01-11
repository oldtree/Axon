package axonx

import (
	"errors"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type SrvConfig struct {
	Port    string `json:"port,omitempty"`
	Host    string `json:"host,omitempty"`
	Version string `json:"version,omitempty"`
	Rate    uint64 `json:"rate,omitempty"`
}

type AxonSrv struct {
	Address  string
	Listener net.Listener

	ReadTimeout       uint64
	WriteTimeout      uint64
	ConnectionTimeout uint64

	SrvExit chan struct{}

	LimitRate int
	Limit     Limiter

	sync.Mutex
	Status   int64
	Inspectx InspectCallBack
}

func NewAxonSrv(i InspectCallBack) *AxonSrv {

	return &AxonSrv{
		Inspectx:     i,
		ReadTimeout:  60,
		WriteTimeout: 60,

		ConnectionTimeout: 10,
		SrvExit:           make(chan struct{}, 1),

		LimitRate: 1000,
		Limit:     NewTimeLimiter(1000),

		Status: 0,
	}
}

func DefaultAxonSrv() *AxonSrv {
	return &AxonSrv{
		ReadTimeout:  60,
		WriteTimeout: 60,

		ConnectionTimeout: 10,
		SrvExit:           make(chan struct{}, 1),
		LimitRate:         0,
		Limit:             NewNoLimiter(),
		Status:            0,
	}
}

func (srv *AxonSrv) IsStart() int64 {
	return atomic.LoadInt64(&srv.Status)
}

func (srv *AxonSrv) Start() error {
	var err error
	srv.Listener, err = net.Listen("tcp", srv.Address)
	if err != nil {
		log.Println("start listener failed : ", err.Error())
		return err
	}
	go srv.Listen()
	time.Sleep(time.Second * 1)
	if atomic.LoadInt64(&srv.Status) != 1 {
		return errors.New("start listener failed ")
	}
	return nil
}

func (srv *AxonSrv) Process(Conn net.Conn) {
	Conn.(*net.TCPConn).SetKeepAlive(true)
	//Conn.(*net.TCPConn).SetDeadline(time.Now().Add(time.Second * 120))
	//Conn.(*net.TCPConn).SetReadDeadline(time.Now().Add(time.Second * 60))
	//Conn.(*net.TCPConn).SetWriteDeadline(time.Now().Add(time.Second * 60))
	cell := NewCellClient(Conn, srv.Inspectx, srv.SrvExit)
	srv.Inspectx.OnConnect(cell)
	go func() {
		defer srv.Inspectx.OnClose(cell)
		cell.Active()
	}()
}

func (srv *AxonSrv) Listen() error {
	defer func() {
		log.Println("listener func is exit : ", time.Now())
		defer atomic.StoreInt64(&srv.Status, 0)
	}()
	atomic.StoreInt64(&srv.Status, 1)
	for {
		select {
		case <-srv.SrvExit:
			log.Println("server is exit : ", time.Now())
		default:
		}

		srv.Limit.Acquire()
		newconn, err := srv.Listener.Accept()
		fi, err := newconn.(*net.TCPConn).File()
		log.Printf("get new connect fd : [%s]  fdptr : [%d] \n", fi.Name(), fi.Fd())
		if err != nil {
			srv.Close()
			log.Println("accept net connect failed : ", err.Error())
			return err
		}
		srv.Process(newconn)
	}
}

func (srv *AxonSrv) Close() {
	err := srv.Listener.Close()
	if err != nil {
		log.Printf("close server listener error : ", err.Error())
	}
	srv.SrvExit <- struct{}{}
	close(srv.SrvExit)
	srv.Limit.Stop()
	atomic.StoreInt64(&srv.Status, 0)
}
