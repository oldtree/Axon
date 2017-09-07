package axonx

import (
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type CellConn struct {
	Conn net.Conn

	Statu bool // connection status

	CloseSignal chan int
	SrvExit     chan int

	ResvChan chan Packet
	SendChan chan Packet

	Inspect InspectCallBack
}

func NewCellClientUseAddress(address string, ins InspectCallBack, srvexit chan int) (*CellConn, error) {
	cn := new(CellConn)
	c, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	cn.Conn = c

	cn.CloseSignal = make(chan int, 1)

	cn.ResvChan = make(chan Packet, 16)
	cn.SendChan = make(chan Packet, 16)
	cn.Inspect = ins
	cn.SrvExit = srvexit
	return cn, nil
}

func NewCellClient(c net.Conn, ins InspectCallBack, srvexit chan int) *CellConn {
	cn := new(CellConn)
	cn.Conn = c
	cn.CloseSignal = make(chan int, 1)

	cn.ResvChan = make(chan Packet, 16)
	cn.SendChan = make(chan Packet, 16)
	cn.Inspect = ins
	cn.SrvExit = srvexit
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
	defer func() {
		if re := recover(); re != nil {
			log.Println("recover panic : ", re)
		}
	}()
	c.Statu = true
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
			break
		case pack, ok := <-c.ResvChan:
			if ok {
				err := c.Inspect.Process(c, pack)
				if err != nil {
					break
				}
			} else {
				log.Println("resv chan is close")
				break
			}
		}
	}
	g.Wait()
	log.Println("sender and revciver is exit")
}

//for top level use
func (c *CellConn) Read(p []byte) (n int, err error) {
	if c.Statu == true {
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
	if c.Statu == true {
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
	if c.Statu == true {
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

func (c *CellConn) WriteWithTimeout(data []byte) (n int, err error) {
	if c.Statu == true {
		packet, err := c.Inspect.Encode(data)
		if err != nil {
			return 0, err
		}
		select {
		case c.SendChan <- packet:
			return packet.Length(), nil
		case <-time.NewTimer(time.Second * 1).C:
			return 0, io.ErrShortBuffer
		}
	} else {
		return 0, errors.New("connection lost")
	}
}

//for real connection send
func (c *CellConn) Send() error {
	defer func() {
		if re := recover(); re != nil {
			log.Println("recover panic : ", re)
		}
	}()
	for {
		select {
		case <-c.CloseSignal:
			log.Printf("cell [%s] close \n", c.RemoteAddr())
			break
		case packSend, ok := <-c.SendChan:
			if ok == true {
				n, err := c.Conn.Write(packSend.GetData())
				if err != nil {
					log.Printf("send data failed [%s] \n", err.Error())
				}
				log.Printf("send data length [%d] \n", n)
			} else {
				log.Println("send chan is close")
			}
		}
	}
	return errors.New("send loop break")
}

//for real connection recive
func (c *CellConn) Recive() error {
	defer func() {
		if re := recover(); re != nil {
			log.Println("recover panic : ", re)
		}
	}()
	//go c.Inspect.SplitMsg(c)
	for {
		select {
		case <-c.CloseSignal:
			log.Printf("cell [%s] close \n", c.RemoteAddr())
			break
		default:
			c.Inspect.SplitMsg(c)
		}

	}
	return errors.New("recive loop break")

}

func (c *CellConn) Close() error {
	err := c.Conn.Close()
	c.CloseSignal <- 1
	close(c.CloseSignal)
	close(c.ResvChan)
	close(c.SendChan)
	return err
}
