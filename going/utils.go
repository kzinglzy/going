package going

import (
	"crypto/sha1"
	"net"
	"strconv"
)

func Sha1(s string) uint64 {
	/*
	   Returns a 160 bit integer based on a
	   string input.
	*/
	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	l := len(bs)
	var a uint64
	for i, b := range bs {
		shift := uint64((l - i - 1) * 8)
		a |= uint64(b) << shift
	}
	return a
}

func GetLocalIP() net.IP {
	addrs, _ := net.InterfaceAddrs()
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP
			}
		}
	}
	return nil
}

func str2uint64(s string) uint64 {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return uint64(n)
}
