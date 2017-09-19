package axonx

import (
	"log"
	"net"
	"sync"
	"time"
)

type SrvConfig struct {
	Port string `json:"port,omitempty"`
	Host string `json:"host,omitempty"`

	Version string `json:"version,omitempty"`
}

type AxonSrv struct {
	Address  string
	Listener net.Listener

	ReadTimeout  uint64
	WriteTimeout uint64

	ConnectionTimeout uint64

	SrvExit chan int

	LimitRate int

	sync.Mutex
	Status bool

	Inspectx InspectCallBack
}

func NewAxonSrv(i InspectCallBack) *AxonSrv {

	return &AxonSrv{
		Inspectx:     i,
		ReadTimeout:  60, //s
		WriteTimeout: 60, //s

		ConnectionTimeout: 10, //s
		SrvExit:           make(chan int, 1),
		LimitRate:         10000, //req/s

		Status: false,
	}
}

func DefaultAxonSrv() *AxonSrv {
	return &AxonSrv{
		ReadTimeout:  60, //s
		WriteTimeout: 60, //s

		ConnectionTimeout: 10, //s
		SrvExit:           make(chan int, 1),
		LimitRate:         10000, //req/s

		Status: false,
	}
}

func (a *AxonSrv) IsStart() bool {
	a.Lock()
	defer a.Unlock()
	return a.Status
}

func (a *AxonSrv) Start() error {
	var err error
	a.Listener, err = net.Listen("tcp", a.Address)
	if err != nil {
		return err
	}
	go a.Listen()
	return nil
}

func (a *AxonSrv) Process(Conn net.Conn) {
	//Conn.SetDeadline(time.Now().Add(time.Second * 60))
	Conn.(*net.TCPConn).SetKeepAlive(true)
	//Conn.SetReadDeadline(time.Now().Add(time.Second * 60))
	//Conn.SetWriteDeadline(time.Now().Add(time.Second * 60))
	cell := NewCellClient(Conn, a.Inspectx, a.SrvExit)
	a.Inspectx.OnConnect(cell)
	go func() {
		defer a.Inspectx.OnClose(cell)
		cell.Active()
	}()
}

func (a *AxonSrv) Listen() error {
	defer func() {
		if re := recover(); re != nil {
			log.Println("recover panic : ", re)
			a.Status = false
		}
	}()
	a.Status = true
	for {
		select {
		case <-a.SrvExit:
			log.Println("server is exit", time.Now())
		default:
		}
		fi, err := a.Listener.(*net.TCPListener).File()
		if err != nil {
			log.Println(fi.Name())
			log.Println(fi.Fd())
		}
		newconn, err := a.Listener.Accept()
		if err != nil {
			log.Println(err)
			a.Close()
			return err
		}
		a.Process(newconn)
	}
}

func (a *AxonSrv) Close() {
	a.SrvExit <- 1
	close(a.SrvExit)
	a.Status = false
	err := a.Listener.Close()
	if err != nil {
		log.Printf("close server listener error : ", err.Error())
	}
}
