package going

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"time"
)

type client struct {
	id         uint64
	c          *net.UDPConn
	peers      map[uint64]*Peer
	localAddr  *net.UDPAddr
	serverAddr *net.UDPAddr
	readQueue  map[uint64]chan *Codec
	writeQueue chan *Codec
	exit       bool
}

func (c *client) Request(method uint16, data []byte, addr *net.UDPAddr) (resp *Response, err error) {
	codec := Codec{
		Method: method,
		Data:   data,
		Addr:   addr,
	}
	reqId := codec.GenRequestId()

	rsponseChan := make(chan *Codec, 1)
	c.readQueue[reqId] = rsponseChan
	defer delete(c.readQueue, reqId)

	c.writeQueue <- &codec

	// wait for response
	select {
	case rp_codec := <-rsponseChan:
		err = json.Unmarshal(rp_codec.Data, &resp)
	case <-time.After(5 * time.Second):
		err = errors.New("client request timeout")
	}
	return
}

func (c *client) Close() {
	c.exit = true
	c.c.Close()
}

func (c *client) registry() error {
	data, _ := json.Marshal(Request{ID: c.id})
	resp, err := c.Request(METHOD_REGISTRY, data, c.serverAddr)
	if err != nil {
		return err
	}
	if resp.Code != CODE_REQUEST_SUCCEED {
		return errors.New("registry failed: " + resp.Body)
	}
	return nil
}

func (c *client) sendLoop() {
	for {
		select {
		case t := <-c.writeQueue:
			c.c.SetWriteDeadline(time.Now().Add(5 * time.Second))
			t.Send(c.c)
		case <-time.After(5 * time.Second):
			log.Println("sending timeout")
			continue
		}
	}
}

func (c *client) readLoop() {
	for !c.exit {
		// parse request
		buf := make([]byte, 65535)
		c.c.SetDeadline(time.Now().Add(time.Second * 5))
		_, addr, err := c.c.ReadFromUDP(buf)
		if err != nil {
			opErr, ok := err.(*net.OpError)
			if ok && opErr.Timeout() {
				continue
			} else {
				log.Print("Client read error:", err, ", Addr:", addr)
				continue
			}
		}
		codec, err := Decode(buf, addr)
		if err != nil {
			log.Print("parse package failed", err)
			continue
		}

		if respChan, ok := c.readQueue[codec.RequestId]; ok {
			respChan <- codec
			continue
		}

		if addr.String() == c.serverAddr.String() {
			// server request
		} else {
			// other client request
		}
	}
}

func (c *client) searchPeers() error {
	resp, err := c.Request(METHOD_GET_PEERS, nil, c.serverAddr)
	if err != nil {
		return err
	}
	if resp.Code != CODE_REQUEST_SUCCEED {
		return errors.New("Get peers failed")
	}
	peers := make(map[string]string)
	if err = json.Unmarshal([]byte(resp.Body), &peers); err != nil {
		return err
	}
	for peerId, udpAddr := range peers {
		pid := str2uint64(peerId)
		udpAddr, _ := net.ResolveUDPAddr("udp", udpAddr)
		c.peers[pid] = &Peer{ID: pid, Addr: udpAddr}
	}
	return nil
}

func NewClient(localAddr string, serverAddr string) (*client, error) {
	c := client{
		id:         NewPeerId(),
		peers:      make(map[uint64]*Peer),
		readQueue:  make(map[uint64]chan *Codec),
		writeQueue: make(chan *Codec, 1000),
		exit:       false,
	}
	c.localAddr, _ = net.ResolveUDPAddr("udp", localAddr)
	c.serverAddr, _ = net.ResolveUDPAddr("udp", serverAddr)

	conn, err := net.ListenUDP("udp", c.localAddr)
	if err != nil {
		return nil, err
	}
	c.c = conn

	go c.readLoop()
	go c.sendLoop()

	err = c.registry()
	if err != nil {
		c.Close()
		return nil, err
	}

	err = c.searchPeers()
	if err != nil {
		c.Close()
		return nil, err
	}
	return &c, nil
}
