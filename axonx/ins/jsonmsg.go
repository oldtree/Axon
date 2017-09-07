package ins

import (
	"bufio"
	"bytes"
	"errors"
	"log"

	"github.com/oldtree/Axon/axonx"
)

var (
	SEPERATE = []byte("\r\n")
)

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
	log.Println(pack.Format())
	n, err := cell.Write([]byte(`{"msg":"hello world"}`))
	if err != nil {
		return err
	}
	log.Println("msg length : ", n)
	return nil
}

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
	}
	return nil

}
