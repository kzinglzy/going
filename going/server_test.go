package going

import (
	"encoding/json"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServerRegistry(t *testing.T) {
	port := ":8888"
	addr := fmt.Sprintf("127.0.0.1%s", port)
	udpAddr, _ := net.ResolveUDPAddr("udp", addr)

	s, err := NewServer(addr)
	require.Nil(t, err)
	conn, err := net.Dial("udp", port)
	require.Nil(t, err)

	rq_id := uint64(123)
	rq := Request{ID: rq_id}
	rq_data := rq.Serialize()
	request := Codec{
		Method:   METHOD_REGISTRY,
		DataSize: uint16(len(rq_data)),
		Data:     rq_data,
	}
	bts, err := request.Encode()
	require.Nil(t, err)
	conn.Write(bts)

	// test response
	data := make([]byte, 1000)
	conn.Read(data)
	codec, _ := Decode(data, udpAddr)
	assert.Equal(t, METHOD_RESPONSE, codec.Method)
	assert.Equal(t, request.RequestId, codec.RequestId)

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	assert.Equal(t, localAddr, s.clients[rq_id])

	var rp Response
	err = json.Unmarshal(codec.Data, &rp)
	require.Nil(t, err)
	assert.Equal(t, CODE_REQUEST_SUCCEED, rp.Code)
	assert.Equal(t, "", rp.Body)

	// test close
	s.Close()
	assert.Equal(t, true, s.exit)

}
