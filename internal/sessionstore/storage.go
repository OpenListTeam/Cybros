package sessionstore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cybros/consts"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
)

const (
	dataDir         = "data"
	sessionFileName = "cybros.session"
)

type encryptedStorage struct {
	path       string
	passphrase string
}

func Init(passphrase string) (telegram.SessionStorage, error) {
	path, err := sessionPath()
	if err != nil {
		return nil, err
	}
	err = ensureDir(path)
	if err != nil {
		return nil, err
	}
	return encryptedStorage{
		path:       path,
		passphrase: passphrase,
	}, nil
}

func (s encryptedStorage) LoadSession(ctx context.Context) ([]byte, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, session.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf(consts.ErrorReadSessionFile, s.path, err)
	}
	if len(data) == 0 {
		return nil, session.ErrNotFound
	}

	file, plaintext, err := decodeSessionFile(data, s.path)
	if err != nil {
		return nil, err
	}
	if plaintext {
		err = s.StoreSession(ctx, data)
		if err != nil {
			return nil, fmt.Errorf(consts.ErrorEncryptPlaintextSessionFile, s.path, err)
		}
		return data, nil
	}

	sessionData, err := decryptSessionData(s.passphrase, file, s.path)
	if err != nil {
		return nil, err
	}
	return sessionData, nil
}

func (s encryptedStorage) StoreSession(_ context.Context, data []byte) error {
	out, err := encryptSessionData(s.passphrase, data)
	if err != nil {
		return err
	}
	err = writeSessionFileAtomic(s.path, out)
	if err != nil {
		return fmt.Errorf(consts.ErrorWriteEncryptedSessionFile, s.path, err)
	}
	return nil
}

func writeSessionFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	tempFile, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-")
	if err != nil {
		return err
	}

	tempPath := tempFile.Name()
	renamed := false
	defer func() {
		if !renamed {
			_ = os.Remove(tempPath)
		}
	}()

	err = tempFile.Chmod(0600)
	if err != nil {
		closeErr := tempFile.Close()
		if closeErr != nil {
			return errors.Join(err, closeErr)
		}
		return err
	}

	_, err = tempFile.Write(data)
	if err != nil {
		closeErr := tempFile.Close()
		if closeErr != nil {
			return errors.Join(err, closeErr)
		}
		return err
	}

	err = tempFile.Close()
	if err != nil {
		return err
	}

	err = os.Rename(tempPath, path)
	if err != nil {
		return err
	}

	renamed = true
	return nil
}

func sessionPath() (string, error) {
	name := filepath.Clean(sessionFileName)
	if name != sessionFileName || name == "." || name == ".." || filepath.IsAbs(name) || filepath.Base(name) != name {
		return "", fmt.Errorf(consts.ErrorInvalidSessionFileName, sessionFileName)
	}

	path := filepath.Join(dataDir, name)
	rel, err := filepath.Rel(dataDir, path)
	if err != nil {
		return "", fmt.Errorf(consts.ErrorCheckSessionPath, path, err)
	}
	if rel == ".." || filepath.IsAbs(rel) || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf(consts.ErrorSessionPathEscapesDataDir, path)
	}

	return path, nil
}

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return errors.New(consts.ErrorSessionDirCurrentDir)
	}

	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return fmt.Errorf(consts.ErrorCreateSessionDir, dir, err)
	}

	info, err := os.Lstat(dir)
	if err != nil {
		return fmt.Errorf(consts.ErrorCheckSessionDir, dir, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf(consts.ErrorSessionDirSymlink, dir)
	}
	if !info.IsDir() {
		return fmt.Errorf(consts.ErrorSessionPathNotDir, dir)
	}

	return nil
}
