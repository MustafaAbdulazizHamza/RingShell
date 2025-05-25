package Pics

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"
)

var timeout time.Duration = 30 * time.Second

func ReceiveAndSaveImage(conn net.Conn, savePath string, imgName string) error {
	reader := bufio.NewReader(conn)

	conn.SetDeadline(time.Now().Add(timeout))
	imgSizeStr, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read image size: %w", err)
	}
	imgSizeStr = imgSizeStr[:len(imgSizeStr)-1] // Remove newline character
	if imgSizeStr == "file is not exist" {
		return os.ErrNotExist
	}
	imgSize, err := strconv.ParseInt(imgSizeStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid image size: %w", err)
	}

	filePath := savePath + "/" + imgName
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()
	conn.SetDeadline(time.Now().Add(timeout))
	_, err = io.CopyN(file, reader, imgSize)
	if err != nil {
		return fmt.Errorf("failed to save image data to file %s: %w", filePath, err)
	}
	return nil
}
