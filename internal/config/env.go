package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"cybros/consts"
)

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf(consts.ErrorOpenFile, path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		after, ok := strings.CutPrefix(line, "export ")
		if ok {
			line = strings.TrimSpace(after)
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf(consts.ErrorDotEnvMissingEqual, path, lineNo)
		}

		key = strings.TrimSpace(key)
		if key == "" || strings.ContainsAny(key, " \t\r\n") {
			return fmt.Errorf(consts.ErrorDotEnvInvalidKey, path, lineNo)
		}

		value = strings.TrimSpace(value)
		unquoted, unquoteErr := strconv.Unquote(value)
		if unquoteErr == nil {
			value = unquoted
		}

		_, exists := os.LookupEnv(key)
		if !exists {
			setErr := os.Setenv(key, value)
			if setErr != nil {
				return fmt.Errorf(consts.ErrorSetEnv, key, setErr)
			}
		}
	}
	scannerErr := scanner.Err()
	if scannerErr != nil {
		return fmt.Errorf(consts.ErrorReadFile, path, scannerErr)
	}
	return nil
}
