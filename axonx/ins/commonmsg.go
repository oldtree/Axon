package ins

import (
	"encoding/binary"
	"errors"
	"hash/crc64"
	"log"
	"sync"
	"sync/atomic"

	"io"

	"fmt"

	"github.com/oldtree/Axon/axonx"
)

var mutex sync.Mutex
var crchash = crc64.New(crc64.MakeTable(crc64.ISO))

type CommonMsg struct {
	Len  uint32 // data + crc
	Data []byte //8 byte
}

func (c *CommonMsg) GetData() []byte {
	return c.Data[4 : c.Len-9]
}

func (c *CommonMsg) WriteData(data []byte) error {
	mutex.Lock()
	defer mutex.Unlock()
	c.Len = uint32(len(data) + 8)
	c.Data = make([]byte, 4+c.Len, 4+c.Len)
	binary.BigEndian.PutUint32(c.Data[0:4], c.Len)

	copy(c.Data[4:c.Len-9], data)
	_, err := crchash.Write(c.Data[4 : c.Len-9])
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint64(c.Data[c.Len-9:], crchash.Sum64())
	crchash.Reset()
	return nil
}
func (c *CommonMsg) Format() string {
	return string(c.Data[4 : c.Len-9])
}
func (c *CommonMsg) Length() int {
	c.Len = binary.BigEndian.Uint32(c.Data[0:4])
	return int(c.Len)
}

type CommonProtcol struct {
}

func (c *CommonProtcol) OnConnect(cell *axonx.CellConn) error {
	log.Println("connection from : ", cell.RemoteAddr().String())
	return nil
}
func (c *CommonProtcol) OnClose(cell *axonx.CellConn) error {
	log.Println("connection close : ", cell.RemoteAddr().String())
	return nil
}

func (c *CommonProtcol) Encode(data []byte) (axonx.Packet, error) {
	var pack = JsonPacket(append(data, SEPERATE...))
	return pack, nil
}
func (c *CommonProtcol) Decode(pack axonx.Packet) ([]byte, error) {
	return []byte(pack.(JsonPacket)), nil
}

func (c *CommonProtcol) Process(cell *axonx.CellConn, pack axonx.Packet) error {
	atomic.AddInt64(&TotalMsgNum, 1)
	log.Printf("msg info : %s number : %d", pack.Format(), TotalMsgNum)
	return nil
}

func (c *CommonProtcol) SplitMsg(cell *axonx.CellConn) (axonx.Packet, error) {
	lengthData := make([]byte, 4, 4)
	var length uint32
	n, err := cell.Conn.Read(lengthData)
	if err != nil || n != 4 {
		return nil, err
	}
	if length = binary.BigEndian.Uint32(lengthData); length > 1024 {
		return nil, errors.New("the size of packet is larger than the limit")
	}
	pack := new(CommonMsg)
	pack.Len = length
	copy(pack.Data[0:4], lengthData)
	pack.Data = make([]byte, length, length)
	n, err = io.ReadFull(cell.Conn, pack.Data[4:length-5])
	if err != nil || n != int(length) {
		return nil, fmt.Errorf("read msg body data failed : %s", err.Error())
	}

	return pack, nil
}
