package main

import (
	"flag"
	"fmt"
	"www.github.com/MustafaAbdulazizHamza/RingShellListener/Art"
	"www.github.com/MustafaAbdulazizHamza/RingShellListener/Server"
)

func main() {
	port := flag.Int("p", 8888, "Port number for the server to listen on")
	https := flag.Bool("https", false, "Enable HTTPS (uses self-signed certificate if no cert/key provided)")
	cert := flag.String("cert", "", "Path to TLS certificate file")
	key := flag.String("key", "", "Path to TLS private key file")
	flag.Parse()

	Art.Art()
	protocol := "HTTP"
	if *https {
		protocol = "HTTPS"
	}
	fmt.Printf("Starting %s server on port %d...\n", protocol, *port)
	Server.Server(*port, *https, *cert, *key)
}
