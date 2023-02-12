package app

import (
	"bufio"
	"encoding/base64"
	"io"
	"os"
)

// EncodeFileB64 encode a file in b64
func EncodeFileB64(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		Logger.Error("unable to open file", err, "file_name", path)
		return "", err
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	content, err := io.ReadAll(reader)
	if err != nil {
		Logger.Error("error wile reading file", err, "file_name", path)
		return "", err
	}
	return base64.StdEncoding.EncodeToString(content), nil
}
