package main

import (
	"net"
)

var ()

func main() {
	raddr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:2308")
	conn, err := net.ListenUDP("udp4", raddr)
	defer conn.Close()

	buf := make([]byte, 1500)
	for {
		//n, addr, err := conn.ReadFromUDP(buf)
		//if err != nil || n == 0 {
		//	continue
		//}
		//packet := buf[:n]
		//
		//if packet == connRequest {
		//conn.WriteToUDP(connResponse, addr)
		//}

	}
}
