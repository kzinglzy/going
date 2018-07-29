package going

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtocol(t *testing.T) {
	content := "test"
	addr := &net.UDPAddr{
		IP:   []byte("127.0.0.1"),
		Port: 7777,
	}
	c := &Codec{
		Method: METHOD_REGISTRY,
		Data:   []byte(content),
		Addr:   addr,
	}

	bts, err := c.Encode()
	if err != nil {
		fmt.Println("err:", err)
	}
	new_codec, err := Decode(bts, addr)
	if err != nil {
		fmt.Println("err:", err)
	}
	assert.Equal(t, c, new_codec, "encode or decode failed")
}
