package files

import (
	"bufio"
	"errors"
	"os"
)

func ReadFile(path string) (string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(contents), nil
}

func ReadFileLines(path string) ([]string, error) {
	fileIn, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer fileIn.Close()

	fileScanner := bufio.NewScanner(fileIn)
	fileScanner.Split(bufio.ScanLines)

	var res []string

	for fileScanner.Scan() {
		res = append(res, fileScanner.Text())
	}
	return res, nil
}

// The following two function oversimplify things
// but this is a toy project so I don't care

func IsDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fileInfo.IsDir()
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}
