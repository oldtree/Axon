package axonx

import (
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type CellConn struct {
	Conn   net.Conn
	Status int64

	CloseSignal chan struct{}
	SrvExit     chan struct{}

	ResvChan chan Packet
	SendChan chan Packet

	Inspect    InspectCallBack
	SingleFunc *sync.Once
}

func NewCellClientUseAddress(address string, ins InspectCallBack, srvexit chan struct{}) (*CellConn, error) {
	cn := new(CellConn)
	c, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	cn.Conn = c

	cn.CloseSignal = make(chan struct{}, 1)

	cn.ResvChan = make(chan Packet, 16)
	cn.SendChan = make(chan Packet, 16)
	cn.Inspect = ins
	cn.SrvExit = srvexit
	cn.SingleFunc = new(sync.Once)
	return cn, nil
}

func NewCellClient(c net.Conn, ins InspectCallBack, srvexit chan struct{}) *CellConn {
	cn := new(CellConn)
	cn.Conn = c
	cn.CloseSignal = make(chan struct{}, 1)

	cn.ResvChan = make(chan Packet, 16)
	cn.SendChan = make(chan Packet, 16)
	cn.Inspect = ins
	cn.SrvExit = srvexit
	cn.SingleFunc = new(sync.Once)
	return cn
}

func (c *CellConn) RemoteAddr() net.Addr {
	addr := c.Conn.RemoteAddr()
	return addr
}

func (c *CellConn) LocalAddr() net.Addr {
	addr := c.Conn.LocalAddr()
	return addr
}

func (c *CellConn) Active() {
	atomic.StoreInt64(&c.Status, 1)
	var g sync.WaitGroup
	go func() {
		g.Add(1)
		defer g.Done()
		c.Send()
	}()
	go func() {
		g.Add(1)
		defer g.Done()
		c.Recive()
	}()
	for {
		select {
		case <-c.SrvExit:
			c.Close()
			goto END
		case <-c.CloseSignal:
			c.Close()
			goto END
		case pack, ok := <-c.ResvChan:
			if ok {
				err := c.Inspect.Process(c, pack)
				if err != nil {
					log.Println("process msg error : ", err.Error())
					goto END
				}
			} else {
				log.Println("resv chan is close")
				goto END
			}
		}
	}
END:
	g.Wait()
	log.Println("sender and revciver is exit")
}

func (c *CellConn) Read(p []byte) (n int, err error) {
	if atomic.LoadInt64(&c.Status) != 1 {
		select {
		case pack, ok := <-c.ResvChan:
			if ok {
				copy(p, pack.GetData())
				return pack.Length(), nil
			} else {
				return 0, io.EOF
			}
		}
	} else {
		return 0, errors.New("connection lost")
	}
}

func (c *CellConn) ReadWithTimeout(p []byte, dur time.Duration) (n int, err error) {
	if atomic.LoadInt64(&c.Status) != 1 {
		select {
		case pack, ok := <-c.ResvChan:
			if ok {
				copy(p, pack.GetData())
				return pack.Length(), nil
			} else {
				return 0, io.EOF
			}
		case <-time.NewTimer(dur).C:
			return 0, io.ErrUnexpectedEOF
		}
	} else {
		return 0, errors.New("connection lost")
	}
}

//for top level use
func (c *CellConn) Write(data []byte) (n int, err error) {
	if atomic.LoadInt64(&c.Status) != 1 {
		packet, err := c.Inspect.Encode(data)
		if err != nil {
			return 0, err
		}
		c.SendChan <- packet
		return packet.Length(), nil
	} else {
		return 0, errors.New("connection lost")
	}
}

func (c *CellConn) WriteWithTimeout(data []byte, dur time.Duration) (n int, err error) {
	if atomic.LoadInt64(&c.Status) != 1 {
		packet, err := c.Inspect.Encode(data)
		if err != nil {
			return 0, err
		}
		select {
		case c.SendChan <- packet:
			return packet.Length(), nil
		case <-time.NewTimer(dur).C:
			return 0, io.ErrShortBuffer
		}
	} else {
		return 0, errors.New("connection lost")
	}
}

//for real connection send
func (c *CellConn) Send() error {
	defer func() {
		log.Printf("cell %s send loop over \n", c.LocalAddr().String())
		c.CloseSignal <- struct{}{}
	}()
	var err error
	for {
		select {
		case <-c.CloseSignal:
			log.Printf("cell [%s] close \n", c.RemoteAddr())
			return nil
		case packSend, ok := <-c.SendChan:
			if ok == true {
				_, err = c.Conn.Write(packSend.GetData())
				if err != nil {
					log.Printf("send data failed [%s] \n", err.Error())
					continue
				}
			} else {
				return errors.New("send chan is close")
			}
		}
	}
	return errors.New("send chan is close")
}

func (c *CellConn) Recive() error {
	defer func() {
		log.Printf("cell %s recive loop over \n", c.LocalAddr().String())
		c.CloseSignal <- struct{}{}
	}()
	for {
		select {
		case <-c.CloseSignal:
			log.Printf("cell [%s] close \n", c.RemoteAddr())
			return nil
		default:
			pack, err := c.Inspect.SplitMsg(c)
			if err != nil {
				log.Println("recive error : ", err.Error())
				continue
			}
			c.ResvChan <- pack
		}

	}
	return errors.New("recive chan close")
}

func (c *CellConn) Close() {
	log.Printf("[%s] is closing \n", c.LocalAddr().String())
	if atomic.LoadInt64(&c.Status) == 1 {
		c.Conn.Close()
		atomic.StoreInt64(&c.Status, 0)
		close(c.CloseSignal)
		close(c.ResvChan)
		close(c.SendChan)
	}
	return
}
