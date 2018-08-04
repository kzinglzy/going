package going

import (
	"encoding/json"
	"fmt"
	"net"
)

type Peer struct {
	ID   uint64       `json:"id"`
	Addr *net.UDPAddr `json:"addr"`
}

func NewPeerId() uint64 {
	ip := GetLocalIP()
	if ip != nil {
		return Sha1(ip.String())
	}
	panic("cant generate peer id")
}

func (p *Peer) Serialize() []byte {
	bytes, err := json.Marshal(p)
	if err != nil {
		panic(fmt.Sprintf("unmarshal object %v %s", p, err))
	}
	return bytes
}

func DeserializePeer(bytes []byte) (*Peer, error) {
	var p = new(Peer)
	err := json.Unmarshal(bytes, p)
	return p, err
}
