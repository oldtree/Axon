package axonx

type Packet interface {
	GetData() []byte
	WriteData([]byte) error
	Format() string
	Length() int
}
