package ins

import (
	"bufio"
	"log"
	"sync/atomic"

	"github.com/oldtree/Axon/axonx"
)

var (
	SEPERATE    = []byte("\r\n")
	TotalMsgNum int64
)

func init() {
	TotalMsgNum = 0
}

type JsonPacket []byte

func (J JsonPacket) GetData() []byte {
	return J
}
func (J JsonPacket) WriteData(data []byte) error {
	J = data
	return nil
}
func (J JsonPacket) Format() string {
	return string(J)
}
func (J JsonPacket) Length() int {
	return len(J)
}

type JsonProtocol struct {
}

func (J *JsonProtocol) OnConnect(cell *axonx.CellConn) error {
	log.Println("connection from : ", cell.RemoteAddr().String())
	return nil
}
func (J *JsonProtocol) OnClose(cell *axonx.CellConn) error {
	log.Println("connection close : ", cell.RemoteAddr().String())
	return nil
}

func (J *JsonProtocol) Encode(data []byte) (axonx.Packet, error) {
	var pack = JsonPacket(append(data, SEPERATE...))
	return pack, nil
}
func (J *JsonProtocol) Decode(pack axonx.Packet) ([]byte, error) {
	return []byte(pack.(JsonPacket)), nil
}

func (J *JsonProtocol) Process(cell *axonx.CellConn, pack axonx.Packet) error {
	atomic.AddInt64(&TotalMsgNum, 1)
	log.Printf("msg info : %s number : %d", pack.Format(), TotalMsgNum)
	return nil
}

func (J *JsonProtocol) SplitMsg(cell *axonx.CellConn) (axonx.Packet, error) {
	reader := bufio.NewReaderSize(cell.Conn, 20)
	data, _, err := reader.ReadLine()
	if err != nil {
		return nil, err
	}
	return JsonPacket(data), nil
}

/*
func (J *JsonProtocol) SplitMsg(cell *axonx.CellConn) error {

	scaner := bufio.NewScanner(cell.Conn)
	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF {
			err = errors.New("EOF")
		}

		index := bytes.Index(data, SEPERATE)
		if index == -1 {
			return
		}
		advance = index + len(SEPERATE)
		token = data[:advance]
		return
	}
	scaner.Split(split)
	for scaner.Scan() {
		data := scaner.Bytes()
		cell.ResvChan <- JsonPacket(data)
		return nil
	}
	return errors.New("empty msg")

}*/
