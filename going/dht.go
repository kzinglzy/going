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

type Message struct {
	Data     []byte `json:"data"`
	FromPeer *Peer  `json:"peer"`
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

func (m *Message) Serialize() []byte {
	bytes, err := json.Marshal(m)
	if err != nil {
		panic(fmt.Sprintf("unmarshal object %v %s", m, err))
	}
	return bytes
}

func DeserializeMessage(bytes []byte) (*Message, error) {
	var m = new(Message)
	err := json.Unmarshal(bytes, m)
	return m, err
}
