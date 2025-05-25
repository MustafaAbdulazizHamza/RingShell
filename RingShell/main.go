package main

import (
	"flag"
	"fmt"
	"www.github.com/MustafaAbdulazizHamza/RingShellListener/Art"
	"www.github.com/MustafaAbdulazizHamza/RingShellListener/Server"
)

func main() {
	port := flag.Int("p", 8888, "Port number for the server to listen on")
	flag.Parse()

	Art.Art()
	fmt.Printf("Starting server on port %d...\n", *port)
	Server.Server(*port)
}
