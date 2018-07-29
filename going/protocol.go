package going

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/sony/sonyflake"
)

// Codec does a encoding and decoding as a transmission protocol
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
	// following data won't be transmited
	Addr *net.UDPAddr
}

type Request struct {
	ID   uint64 `json:"id"`   // request peer id
	Body string `json:"body"` // request content
}

type Response struct {
	Code uint16 `json:"code"`
	Body string `json:"body"`
}

// communicate method
const (
	// client to server
	METHOD_RESPONSE    = uint16(1)
	METHOD_REGISTRY    = uint16(2)
	METHOD_GET_PEERS   = uint16(3) // TODO, is this required
	METHOD_SEARCH_PEER = uint16(4)

	// client to client
	METHOD_SEND_MESSAGE = uint16(1000)
)

// response code
const (
	CODE_REQUEST_SUCCEED       = uint16(1)
	CODE_INVALID_PARAM         = uint16(2)
	CODE_INTERNAL_SERVER_ERROR = uint16(3)
	CODE_NOT_FOUND             = uint16(4)
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

// unique id generator by snoyflake algorithm
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

func (r *Request) Serialize() []byte {
	bytes, err := json.Marshal(r)
	if err != nil {
		panic(fmt.Sprintf("unmarshal object %v %s", r, err))
	}
	return bytes
}

func DeserializeRequest(bytes []byte) (*Request, error) {
	var req = new(Request)
	err := json.Unmarshal(bytes, req)
	return req, err
}
