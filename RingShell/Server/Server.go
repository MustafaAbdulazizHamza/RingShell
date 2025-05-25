package Server

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

var (
	address                    string = ""
	port                       int    = 8642
	dt                                = make(map[string][]string)
	chs                               = make(map[string]chan string)
	mu                         sync.RWMutex
	CId                        string = ""
	outputDir                  string
	timeout                    time.Duration = 5 * time.Second
	defaultPortNumber          string        = "65000"
	controllingServersChannels               = make(map[string]chan string)
	listeningServersChannels                 = make(map[string]chan string)
	controllingServers                       = make(map[string]string)
	listeningServers                         = make(map[string]string)
)

func Server(port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		log.Fatalln("Unable to bind to port:", err)
	}
	defer listener.Close()
	go func() {
		i := 0
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("Unable to accept connection:", err)
				continue
			}
			d, err := getUserInfo(conn)
			if err != nil {
				log.Println("Unable to fetch user`s information:", err)
				continue
			}
			mu.Lock()
			if _, exists := dt[fmt.Sprint(i)]; exists {
				conn.Close()
				continue
			}
			dt[fmt.Sprint(i)] = d
			chs[fmt.Sprint(i)] = make(chan string)
			mu.Unlock()
			go handle(conn, fmt.Sprint(i))
			i += 1

		}
	}()
	if dir, err := os.Getwd(); err != nil {
		log.Println("Unable to get the Current Working Directory:", err)
	} else {
		outputDir = dir
	}
	inputParser()
}
