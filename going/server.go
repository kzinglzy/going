package going

import (
	"encoding/json"
	"log"
	"net"
	"time"
)

type server struct {
	c          *net.UDPConn            // current connection
	exit       bool                    // is server exit
	clients    map[uint64]*net.UDPAddr // connected clients
	writeQueue chan *Codec             // to send msg
}

func (s *server) listen() {
	for !s.exit {
		buf := make([]byte, 65535)
		s.c.SetDeadline(time.Now().Add(time.Second * 5))
		_, addr, err := s.c.ReadFromUDP(buf)
		if err != nil {
			opErr, ok := err.(*net.OpError)
			if ok && opErr.Timeout() {
				continue
			} else {
				log.Print("Server read error:", err, ", Addr:", addr)
				continue
			}
		}

		codec, err := Decode(buf, addr)
		if err != nil {
			log.Print("parse package failed", err)
		}
		go s.handle(codec)
	}
}

func (s *server) response(method uint16, code uint16, body string, requestId uint64, addr *net.UDPAddr) *Codec {
	data, _ := json.Marshal(Response{Code: code, Body: body})
	resp := Codec{
		Method:    method,
		Data:      data,
		Addr:      addr,
		RequestId: requestId,
	}
	s.writeQueue <- &resp
	return &resp
}

func (s *server) handle(c *Codec) {
	var body string

	switch c.Method {
	case METHOD_REGISTRY:
		req, err := DeserializeRequest(c.Data)
		if err != nil {
			body = "invalid registry reqeust body" + err.Error()
			s.response(METHOD_RESPONSE, CODE_INVALID_PARAM, body, c.RequestId, c.Addr)
		} else {
			s.clients[req.ID] = c.Addr
			log.Println("client registry ", c.Addr, req.ID)
			s.response(METHOD_RESPONSE, CODE_REQUEST_SUCCEED, body, c.RequestId, c.Addr)
		}
	case METHOD_GET_PEERS:
		addrs := make(map[uint64]string)
		for peerId, udpAddr := range s.clients {
			addrs[peerId] = udpAddr.String()
		}
		bytes, err := json.Marshal(addrs)
		if err != nil {
			s.response(METHOD_RESPONSE, CODE_INTERNAL_SERVER_ERROR, "marshal peers failed", c.RequestId, c.Addr)
		} else {
			body = string(bytes)
			s.response(METHOD_RESPONSE, CODE_REQUEST_SUCCEED, body, c.RequestId, c.Addr)
		}
	case METHOD_SEARCH_PEER:
		req, err := DeserializeRequest(c.Data)
		if err != nil {
			body = "invalid search peer body" + err.Error()
			s.response(METHOD_RESPONSE, CODE_INVALID_PARAM, body, c.RequestId, c.Addr)
			return
		}
		peer, err := DeserializePeer([]byte(req.Body))
		if err != nil {
			s.response(METHOD_RESPONSE, CODE_INVALID_PARAM, body, c.RequestId, c.Addr)
			return
		}
		peerUDPAddr, found := s.clients[peer.ID]
		if !found {
			body = "cant found corresponding peer"
			s.response(METHOD_RESPONSE, CODE_NOT_FOUND, body, c.RequestId, c.Addr)
		} else {
			peer.Addr = peerUDPAddr
			body = string(peer.Serialize())
			s.response(METHOD_RESPONSE, CODE_REQUEST_SUCCEED, body, c.RequestId, c.Addr)
		}
	}
}

func (s *server) sendLoop() {
	for !s.exit {
		select {
		case c := <-s.writeQueue:
			s.c.SetWriteDeadline(time.Now().Add(5 * time.Second))
			c.Send(s.c)
		case <-time.After(5 * time.Second):
			log.Println("sending msg timeout")
			continue
		}
	}
}

func (s *server) Close() {
	s.exit = true
	s.c.Close()
}

func NewServer(address string) (*server, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}

	c, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	s := &server{
		c:          c,
		exit:       false,
		clients:    make(map[uint64]*net.UDPAddr),
		writeQueue: make(chan *Codec, 1000),
	}
	go s.listen()
	go s.sendLoop()
	return s, nil
}
