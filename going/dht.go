package going

import "net"

type Peer struct {
	ID   uint64
	Addr *net.UDPAddr
}

func NewPeerId() uint64 {
	ip := GetLocalIP()
	if ip != nil {
		return Sha1(ip.String())
	}
	panic("cant generate peer id")
}
