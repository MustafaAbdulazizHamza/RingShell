package Server

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
)

func bindPort(cmd []string) {
	if len(cmd) < 4 {
		log.Println("Invalid bind command format. Usage: bind <type> <port> named <name>")
		return
	}

	port := cmd[1]
	if !isAValidPortNumber(cmd[1]) {
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
		return
	}

	chann := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())

	mu.Lock()
	listeningServers[name] = port
	listeningServersChannels[name] = chann
	mu.Unlock()

	outDir := filepath.Join(outputDir, name)
	if !isExist(outDir) {
		os.MkdirAll(outDir, 0755)
	}

	connectionCounter := 0
	var counterMu sync.Mutex

	go func() {
		defer listener.Close()
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
					continue
				}
				counterMu.Lock()
				currentID := connectionCounter
				connectionCounter++
				counterMu.Unlock()
				go listenerServiceHandler(ctx, conn, outDir, name, currentID)
			}
		}
	}()

	for {
		message := <-chann
		if message == "exit" {
			cancel()
			mu.Lock()
			delete(listeningServersChannels, name)
			delete(listeningServers, name)
			mu.Unlock()
			break
		}
	}
}

func listenerServiceHandler(ctx context.Context, conn net.Conn, outDir string, outFile string, id int) error {
	defer conn.Close()

	fi, err := os.OpenFile(filepath.Join(outDir, fmt.Sprintf("%s-%d.txt", outFile, id)), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer fi.Close()

	done := make(chan error, 1)
	go func() {
		_, err := io.Copy(fi, conn)
		done <- err
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-done:
		return err
	}
}

func setUpControllerServer(port string, name string) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", address, port))
	if err != nil {
		log.Println("Unable to bind to port", port, ".")
		return
	}

	zombiesChannels := make([]chan string, 0)
	var zombiesChannelsMu sync.Mutex
	chann := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())

	mu.Lock()
	controllingServersChannels[name] = chann
	controllingServers[name] = port
	mu.Unlock()

	go func() {
		defer listener.Close()
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
					continue
				}
				zombiesCh := make(chan string, 10)
				zombiesChannelsMu.Lock()
				zombiesChannels = append(zombiesChannels, zombiesCh)
				zombiesChannelsMu.Unlock()
				go zombiesHandler(conn, zombiesCh)
			}
		}
	}()

	for {
		message, open := <-chann
		if !open {
			break
		}

		zombiesChannelsMu.Lock()
		for _, ch := range zombiesChannels {
			select {
			case ch <- message:
			default:
			}
		}
		zombiesChannelsMu.Unlock()

		if message == "exit" {
			cancel()
			break
		}
	}

	mu.Lock()
	delete(controllingServersChannels, name)
	delete(controllingServers, name)
	mu.Unlock()
}

func zombiesHandler(conn net.Conn, chann chan string) {
	defer conn.Close()
	writer := bufio.NewWriter(conn)

	for {
		command, open := <-chann
		if !open || command == "exit" {
			return
		}

		if _, err := writer.WriteString(command + "\n"); err != nil {
			return
		}
		if err := writer.Flush(); err != nil {
			return
		}
	}
}
