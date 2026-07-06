package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"cybros/consts"
	"cybros/internal/logger"
)

const (
	dataDir  = "data"
	cacheDir = "cache"
)

type fileEntry[T any] struct {
	Timestamp int64 `json:"Timestamp"`
	Value     T     `json:"Value"`
}

var (
	lockMutex sync.Mutex
	fileLocks = map[string]*sync.Mutex{}
)

func Get[T any](namespace string, key string, ttl time.Duration) (T, bool) {
	var zero T

	path := cachePath(namespace, key)
	unlock := lockFile(path)
	defer unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return zero, false
	}

	var entry fileEntry[T]
	err = json.Unmarshal(data, &entry)
	if err != nil {
		return zero, false
	}
	if entry.Timestamp <= 0 {
		return zero, false
	}
	if expired(entry.Timestamp, ttl) {
		return zero, false
	}

	return entry.Value, true
}

func Set[T any](namespace string, key string, value T) {
	path := cachePath(namespace, key)
	unlock := lockFile(path)
	defer unlock()

	err := write(path, value)
	if err == nil {
		return
	}

	err = write(path, value)
	if err != nil {
		logWriteError(path, err)
	}
}

func Load[T any](namespace string, key string, ttl time.Duration, loader func() (T, error)) (T, error) {
	value, ok := Get[T](namespace, key, ttl)
	if ok {
		return value, nil
	}

	value, err := loader()
	if err != nil {
		var zero T
		return zero, err
	}

	Set(namespace, key, value)
	return value, nil
}

func cachePath(namespace string, key string) string {
	hash := sha256.Sum256([]byte(key))
	fileName := hex.EncodeToString(hash[:]) + ".json"
	return filepath.Join(dataDir, cacheDir, safeName(namespace), fileName)
}

func lockFile(path string) func() {
	lockMutex.Lock()
	fileMutex := fileLocks[path]
	if fileMutex == nil {
		fileMutex = &sync.Mutex{}
		fileLocks[path] = fileMutex
	}
	lockMutex.Unlock()

	fileMutex.Lock()
	return fileMutex.Unlock
}

func safeName(value string) string {
	var builder strings.Builder
	for _, r := range value {
		if r >= 'a' && r <= 'z' {
			builder.WriteRune(r)
			continue
		}
		if r >= 'A' && r <= 'Z' {
			builder.WriteRune(r)
			continue
		}
		if r >= '0' && r <= '9' {
			builder.WriteRune(r)
			continue
		}
		if r == '_' || r == '-' {
			builder.WriteRune(r)
			continue
		}
		builder.WriteByte('_')
	}
	if builder.Len() == 0 {
		return "default"
	}
	return builder.String()
}

func expired(timestamp int64, ttl time.Duration) bool {
	if ttl <= 0 {
		return false
	}
	return time.Now().Unix()-timestamp > int64(ttl.Seconds())
}

func write[T any](path string, value T) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	entry := fileEntry[T]{
		Timestamp: time.Now().Unix(),
		Value:     value,
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}

	tempPath := path + ".tmp"
	err = os.WriteFile(tempPath, data, 0600)
	if err != nil {
		return err
	}

	err = os.Rename(tempPath, path)
	if err != nil {
		return err
	}
	return nil
}

func logWriteError(path string, err error) {
	if logger.Log == nil {
		return
	}
	logger.Log.WithError(err).WithField("path", path).Error(consts.ErrorWriteCacheFile)
}
