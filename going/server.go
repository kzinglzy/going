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

type Request struct {
	ID uint64 `json:"id"`
}

type Response struct {
	Code uint16 `json:"code"`
	Body string `json:"body"`
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
	return &resp
}

func (s *server) handle(c *Codec) {
	log.Printf("handle msg from %s, method %d\n", c.Addr, c.Method)
	var (
		code uint16 = CODE_REQUEST_SUCCEED
		body string
	)
	switch c.Method {
	case METHOD_REGISTRY:
		var req Request
		err := json.Unmarshal(c.Data, &req)
		if err != nil {
			code = CODE_INVALID_PARAM
			body = "invalid registry reqeust body" + err.Error()

		} else {
			s.clients[req.ID] = c.Addr
			log.Println("registry ", c.Addr, req.ID)
		}
		s.writeQueue <- s.response(METHOD_REGISTRY_OK, code, body, c.RequestId, c.Addr)
	case METHOD_GET_PEERS:
		addrs := make(map[uint64]string)
		for peerId, udpAddr := range s.clients {
			addrs[peerId] = udpAddr.String()
		}
		bytes, err := json.Marshal(addrs)
		if err != nil {
			log.Println("json marshal peers failed ", err)
			return
		}
		body = string(bytes)
		s.writeQueue <- s.response(METHOD_GET_PEERS_OK, CODE_REQUEST_SUCCEED, body, c.RequestId, c.Addr)
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
