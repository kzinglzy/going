package going

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/sony/sonyflake"
)

// Codec does a encoding and decoding as a commuticate protocol
// The data format:
//  | header                                  | body
//  | uint16         | uint64     | uint16    | bytes |
//  | request method | requsts id | data size | data  |
var CODEC_HEADER_LEN uint16 = 12

type Codec struct {
	Method    uint16
	RequestId uint64
	DataSize  uint16
	Data      []byte
	Addr      *net.UDPAddr // 不参与网络传送
}

// communicate method
const (
	METHOD_REGISTRY     = uint16(1)
	METHOD_REGISTRY_OK  = uint16(2)
	METHOD_GET_PEERS    = uint16(3)
	METHOD_GET_PEERS_OK = uint16(4)
)

// response code
const (
	CODE_REQUEST_SUCCEED = uint16(1)
	CODE_INVALID_PARAM   = uint16(2)
)

func (c *Codec) Send(conn *net.UDPConn) {
	bts, err := c.Encode()
	if err != nil {
		log.Println("encode failed", err)
	}
	_, err = conn.WriteToUDP(bts, c.Addr)
	if err != nil {
		log.Println("write failed", err)
	}
}

var sf *sonyflake.Sonyflake = sonyflake.NewSonyflake(sonyflake.Settings{})

func (c *Codec) GenRequestId() uint64 {
	if c.RequestId == 0 {
		id, _ := sf.NextID()
		c.RequestId = id
	}
	return c.RequestId
}

func (c *Codec) Encode() ([]byte, error) {
	w := new(bytes.Buffer)

	if err := binary.Write(w, binary.BigEndian, c.Method); err != nil {
		return nil, err
	}

	c.GenRequestId()
	if err := binary.Write(w, binary.BigEndian, c.RequestId); err != nil {
		return nil, err
	}

	c.DataSize = uint16(len(c.Data))
	if err := binary.Write(w, binary.BigEndian, c.DataSize); err != nil {
		return nil, err
	}

	n, err := w.Write(c.Data)
	if err != nil {
		return nil, err
	}
	if n != int(c.DataSize) {
		return nil, errors.New(fmt.Sprintf("Data size not match, excepted %d, actualy %d", c.DataSize, n))
	}
	return w.Bytes(), nil
}

func Decode(bts []byte, addr *net.UDPAddr) (*Codec, error) {
	r := bytes.NewBuffer(bts)
	c := &Codec{}

	if err := binary.Read(r, binary.BigEndian, &(c.Method)); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.BigEndian, &(c.RequestId)); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.BigEndian, &(c.DataSize)); err != nil {
		return nil, err
	}
	c.Data = make([]byte, c.DataSize)
	copy(c.Data, bts[CODEC_HEADER_LEN:])
	c.Addr = addr
	return c, nil
}
