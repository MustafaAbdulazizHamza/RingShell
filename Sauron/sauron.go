package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
)

var (
	ipFlag   string
	portFlag int
	outFlag  string
	opS      string
	arch     string
)

func init() {
	flag.StringVar(&ipFlag, "ip", "", "Listener IP address")
	flag.IntVar(&portFlag, "port", 0, "Listener port number")
	flag.StringVar(&outFlag, "out", "", "Output directory (Full path)")
	flag.StringVar(&opS, "os", "", "The target operating system (lowercase)")
	flag.StringVar(&arch, "arch", "", "The target architecture")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage of %s:
A tool that is used to generate rings based on user input.

Flags:
`, os.Args[0])
		flag.PrintDefaults()

		fmt.Fprintln(os.Stderr, `
Example:
  ./sauron -ip 192.168.1.10 -port 8080 -os windows -arch amd64 -out <output path>`)
	}
}

func main() {
	flag.Parse()

	if ipFlag == "" || portFlag == 0 || outFlag == "" || opS == "" || arch == "" {
		fmt.Fprint(os.Stderr, "Error: all flags --ip, --port, --os, --arch, and --out are required.\n\n")
		flag.Usage()
		os.Exit(1)
	}

	outputFilePath := "src/add.go"
	file, err := os.OpenFile(outputFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("Failed to open file %s: %v", outputFilePath, err)
	}
	defer file.Close()

	content := fmt.Sprintf(`package main

var (
	address string = "%s"
	port    int    = %d
)
`, ipFlag, portFlag)

	if _, err := file.WriteString(content); err != nil {
		log.Fatalf("Failed to write to file: %v", err)
	}
	if err = callBuildScript(opS, arch, outFlag); err != nil {
		log.Fatal("Unable to call the build script: ", err)
	}
}

func callBuildScript(goos, goarch, outDir string) error {
	scriptPath := "./build.sh"

	cmd := exec.Command(scriptPath, goos, goarch, outDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Building failed: %w", err)
	}

	return nil
}
