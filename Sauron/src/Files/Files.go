package Files

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func SendFileContents(filePath string, conn net.Conn, timeout time.Duration) error {
	writer := bufio.NewWriter(conn)
	fi, err := os.Open(filePath)
	if err != nil {
		if _, err := fmt.Fprint(writer, "file is not exist"); err != nil {
			return err
		}
		return err
	}
	defer fi.Close()
	contents, err := io.ReadAll(fi)
	if err != nil {
		return err
	}

	conn.SetDeadline(time.Now().Add(timeout))
	if _, err := fmt.Fprintf(writer, "%d\n", len(contents)); err != nil {
		return err
	}
	conn.SetDeadline(time.Now().Add(timeout))
	if err := write(contents, writer); err != nil {
		return err
	}

	return nil
}

func ReceiveFileContents(fileName string, conn net.Conn, timeout time.Duration) error {
	fi, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer fi.Close()
	reader := bufio.NewReader(conn)
	conn.SetDeadline(time.Now().Add(timeout))
	size, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	if size == "file is not exist" {
		return os.ErrNotExist
	}
	fileSize, err := strconv.Atoi(strings.TrimSpace(size))
	if err != nil {
		return err
	}
	contents := make([]byte, fileSize)
	conn.SetDeadline(time.Now().Add(timeout))
	if _, err := io.ReadFull(reader, contents); err != nil {
		return err
	}
	if _, err := fi.Write(contents); err != nil {
		return err
	}
	return nil
}

func write(data []byte, writer *bufio.Writer) error {
	if _, err := writer.Write(data); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	return nil
}
