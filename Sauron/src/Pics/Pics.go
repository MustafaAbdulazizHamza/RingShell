package Pics

import (
	"bufio"
	"fmt"
	"image/png"
	"io"
	"net"
	"os"

	"github.com/kbinani/screenshot"
)

func SendScreenshots(conn net.Conn) error {
	writer := bufio.NewWriter(conn)
	numScreenshots := screenshot.NumActiveDisplays()
	if _, err := writer.WriteString(fmt.Sprintf("%d\n", numScreenshots)); err != nil {
		return fmt.Errorf("failed to send number of screenshots: %w", err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	for i := 0; i < numScreenshots; i++ {
		img, err := screenshot.CaptureDisplay(i)
		if err != nil {
			return fmt.Errorf("failed to capture screenshot: %w", err)
		}

		imgFileName := fmt.Sprintf("screenshot_%d.png", i)
		imgFile, err := os.Create(imgFileName)
		if err != nil {
			return fmt.Errorf("failed to create screenshot file: %w", err)
		}

		if err := png.Encode(imgFile, img); err != nil {
			imgFile.Close()
			return fmt.Errorf("failed to save screenshot to file: %w", err)
		}

		imgFile.Close()

		if err := sendImageFile(conn, imgFileName); err != nil {
			return fmt.Errorf("failed to send screenshot %s: %w", imgFileName, err)
		}
		os.Remove(imgFileName)
	}

	return nil
}

func SendImageFromPath(conn net.Conn, imagePath string) error {
	return sendImageFile(conn, imagePath)
}

func sendImageFile(conn net.Conn, filePath string) error {
	writer := bufio.NewWriter(conn)

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	imgSize := fileInfo.Size()

	if _, err := writer.WriteString(fmt.Sprintf("%d\n", imgSize)); err != nil {
		return fmt.Errorf("failed to send image size: %w", err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	imgFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open image file: %w", err)
	}
	defer imgFile.Close()

	if _, err := io.Copy(writer, imgFile); err != nil {
		return fmt.Errorf("failed to send image data: %w", err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer after sending image: %w", err)
	}

	return nil
}
