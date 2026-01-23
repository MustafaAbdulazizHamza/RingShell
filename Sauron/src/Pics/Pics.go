package Pics

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"io"
	"os"
	"strings"

	"github.com/kbinani/screenshot"
)

func GetScreenshotsData() (string, error) {
	numScreenshots := screenshot.NumActiveDisplays()
	var results []string

	for i := 0; i < numScreenshots; i++ {
		img, err := screenshot.CaptureDisplay(i)
		if err != nil {
			return "", fmt.Errorf("failed to capture screenshot: %w", err)
		}

		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return "", fmt.Errorf("failed to encode screenshot: %w", err)
		}

		encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
		results = append(results, fmt.Sprintf("screenshot_%d:%s", i, encoded))
	}

	return strings.Join(results, "\n"), nil
}

func GetImageData(imagePath string) (string, error) {
	fi, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer fi.Close()

	data, err := io.ReadAll(fi)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}
