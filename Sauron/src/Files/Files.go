package Files

import (
	"encoding/base64"
	"io"
	"os"
)

func GetFileData(filePath string) (string, error) {
	fi, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer fi.Close()

	contents, err := io.ReadAll(fi)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(contents), nil
}
