package going

import (
	"encoding/json"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServerRegistry(t *testing.T) {
	port := ":8888"
	addr := fmt.Sprintf("127.0.0.1%s", port)
	udpAddr, _ := net.ResolveUDPAddr("udp", addr)

	s, err := NewServer(addr)
	if err != nil {
		t.Error(err)
	}
	conn, err := net.Dial("udp", port)
	if err != nil {
		t.Error(err)
	}

	rq_id := uint64(123)
	rq_data, _ := json.Marshal(Request{ID: rq_id})
	request := Codec{
		Method:   METHOD_REGISTRY,
		DataSize: uint16(len(rq_data)),
		Data:     rq_data,
	}
	bts, err := request.Encode()
	if err != nil {
		t.Error(err)
	}
	conn.Write(bts)

	// test response
	data := make([]byte, 1000)
	conn.Read(data)
	codec, _ := Decode(data, udpAddr)
	assert.Equal(t, METHOD_REGISTRY_OK, codec.Method)
	assert.Equal(t, request.RequestId, codec.RequestId)

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	assert.Equal(t, localAddr, s.clients[rq_id])

	var rp Response
	err = json.Unmarshal(codec.Data, &rp)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, CODE_REQUEST_SUCCEED, rp.Code)
	assert.Equal(t, "", rp.Body)

	// test close
	s.Close()
	assert.Equal(t, true, s.exit)

}
