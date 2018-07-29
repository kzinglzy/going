package going

import (
	"fmt"
	"testing"
)

var serverAddr = "127.0.0.1:8887"

func init() {
	_, err := NewServer(serverAddr)
	if err != nil {
		panic("cannt start server")
	}
}

func TestClient(t *testing.T) {
	c, err := NewClient("127.0.0.1:6666", serverAddr)
	if err != nil {
		t.Error(err)
	}
	for id, addr := range c.peers {
		fmt.Println(id, addr.Addr, "++++++++")
	}

	// assert.Equal(t, c.)
}
