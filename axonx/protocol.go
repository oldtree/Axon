package axonx

type Guarder interface {
	OnConnect(cell *CellConn) error
	OnClose(cell *CellConn) error
	//OnConnectiontStat(cell *CellConn) error
}

type InspectCallBack interface {
	Guarder
	SplitMsg(cell *CellConn) error
	Process(cell *CellConn, pack Packet) error
	Encode([]byte) (Packet, error)
	Decode(Packet) ([]byte, error)
}
