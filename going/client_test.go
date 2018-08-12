package going

import (
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/bouk/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var serverAddr string = "127.0.0.1:8887"
var localAddressA string = "127.0.0.1:6666"
var localAddressB string = "127.0.0.1:6667"

func init() {
	_, err := NewServer(serverAddr)
	if err != nil {
		panic("cant create server")
	}
}

type DefaultHandler struct{}
type ResponseHandler struct{}

func (h *DefaultHandler) HandleData(c *Conn) {
	m := c.GetMessage()
	log.Println(fmt.Sprintf("%s>[%s]Say: %s", c.localAddr, m.FromPeer.Addr, m.Data))
}

func (h *ResponseHandler) HandleData(c *Conn) {
	m := c.GetMessage()
	log.Println(fmt.Sprintf("%s>[%s]Say: %s", c.localAddr, m.FromPeer.Addr, m.Data))
	c.SendMessage(m.FromPeer.ID, "pong!")
}

var defaultHandler Handler = new(DefaultHandler)
var responseHandler Handler = new(ResponseHandler)

// mockNewPeerId patch NewPeerId to simulate mulit clients
func mockNewPeerId() func() {
	monkey.Patch(NewPeerId, func() uint64 {
		return rand.Uint64()
	})
	return func() {
		monkey.Unpatch(NewPeerId)
	}
}

func TestNewClient(t *testing.T) {
	ca, err := NewClient(localAddressA, serverAddr, defaultHandler)
	defer ca.Close()
	require.Nil(t, err)
	NewClient(localAddressA, serverAddr, defaultHandler) // same ip will considerd to be the same client
	assert.Equal(t, 1, len(ca.peers))
	for _, peer := range ca.peers {
		assert.Equal(t, localAddressA, peer.Addr.String(), "invalid address")
	}

	defer mockNewPeerId()()

	cb, err := NewClient(localAddressB, serverAddr, defaultHandler)
	defer cb.Close()
	require.Nil(t, err)
	assert.Equal(t, 2, len(cb.peers))
	var addrs []string
	for _, peer := range cb.peers {
		addrs = append(addrs, peer.Addr.String())
	}
	assert.ElementsMatch(t, []string{localAddressA, localAddressB}, addrs)
}

func TestClientDialPeer(t *testing.T) {
	defer mockNewPeerId()()
	ca, err := NewClient(localAddressA, serverAddr, defaultHandler)
	require.Nil(t, err)
	defer ca.Close()
	cb, err := NewClient(localAddressB, serverAddr, defaultHandler)
	require.Nil(t, err)
	defer cb.Close()

	_, found := ca.peers[cb.id]
	assert.False(t, found)

	peer, err := ca.dialPeer(cb.id)
	require.Nil(t, err)
	assert.Equal(t, cb.id, peer.ID)
	assert.Equal(t, cb.localAddr, peer.Addr)
	_, found = ca.peers[cb.id]
	assert.True(t, found)
}

func TestClientSendingMessage(t *testing.T) {
	defer mockNewPeerId()()
	ca, err := NewClient(localAddressA, serverAddr, defaultHandler)
	require.Nil(t, err)
	defer ca.Close()

	cb, err := NewClient(localAddressB, serverAddr, responseHandler)
	require.Nil(t, err)
	defer cb.Close()

	err = ca.SendMessage(cb.id, "ping!")
	require.Nil(t, err)
	time.Sleep(time.Second * 1)
}
