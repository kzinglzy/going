package going

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
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
	require.Nil(t, err)
	new_codec, err := Decode(bts, addr)
	require.Nil(t, err)
	require.Equal(t, c, new_codec, "encode or decode failed")
}
