package going

import (
	"math/rand"
	"testing"

	"github.com/bouk/monkey"
	"github.com/stretchr/testify/assert"
)

var serverAddr = "127.0.0.1:8887"

func init() {
	_, err := NewServer(serverAddr)
	if err != nil {
		panic("cannt start server")
	}
}

func TestNewClient(t *testing.T) {
	assert := assert.New(t)

	address := "127.0.0.1:6666"
	new_address := "127.0.0.1:6667"

	c, err := NewClient(address, serverAddr)
	if err != nil {
		t.Error(err)
	}
	NewClient(address, serverAddr) // same ip will considerd to be the same client
	assert.Equal(1, len(c.peers))
	for _, peer := range c.peers {
		assert.Equal(address, peer.Addr.String(), "invalid address")
	}

	// patch NewPeerId to simulate mulit clients
	monkey.Patch(NewPeerId, func() uint64 {
		return rand.Uint64()
	})
	defer monkey.Unpatch(NewPeerId)

	new_c, err := NewClient(new_address, serverAddr)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(2, len(new_c.peers))
	var addrs []string
	for _, peer := range new_c.peers {
		addrs = append(addrs, peer.Addr.String())
	}
	assert.ElementsMatch([]string{address, new_address}, addrs)
}
