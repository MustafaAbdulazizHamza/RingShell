package Server

import (
	"bufio"
	"log"
	"os"
	"strings"
)

func importScript(script string) {
	scr, err := os.Open(script)
	if err != nil {
		log.Println("Unable to open the script file:", err)
		return
	}
	defer scr.Close()
	scanner := bufio.NewScanner(scr)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "#") {
			s := strings.Split(line, "#")
			line = s[0]
		}
		line = strings.TrimSpace(line)
		if line != "" {
			executor(line)
		}
	}
}
