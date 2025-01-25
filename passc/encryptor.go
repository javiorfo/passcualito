package passc

import (
	"errors"
	"fmt"
	"strings"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"os"
	"path/filepath"

	"github.com/javiorfo/steams/opt"
)

type Encryptor struct {
	MasterPassword string
	FilePath       string
}

func (e Encryptor) EncryptText(text string) error {
	file, err := os.OpenFile(e.FilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	finalText := text
	stat, _ := file.Stat()
	if stat.Size() != 0 {
		decryptedText, err := e.ReadEncryptedText()
		if err != nil {
			return err
		}
		finalText += ";" + decryptedText
	}

	block, err := aes.NewCipher([]byte(e.MasterPassword))
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(finalText), nil)

	if _, err := file.Write(append(nonce, ciphertext...)); err != nil {
		return err
	}

	return nil
}

func (e Encryptor) ReadEncryptedText() (string, error) {
	file, err := os.Open(e.FilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	block, err := aes.NewCipher([]byte(e.MasterPassword))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	nonce := data[:gcm.NonceSize()]
	if len(data) == 0 {
		return "", errors.New("No passwords stored")
	}
	ciphertext := data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func newTemp(masterPassword string) error {
	key, err := generateRandomPassword(16)
	filePath := fmt.Sprintf("/tmp/%s.passc", *key)
	if err != nil {
		return fmt.Errorf("Error generating tmp password: %v", err)
	}
	encryptor := &Encryptor{MasterPassword: *key, FilePath: filePath}
	encryptor.EncryptText(masterPassword)
	return nil
}

func getTempEncryptor(filePath string) opt.Optional[Encryptor] {
	var tempFilePath string

	err := filepath.Walk("/tmp", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(info.Name()) == passcExtension {
			tempFilePath = path
			return fmt.Errorf("NO_PASSC")
		}

		return nil
	})

	if err != nil && err.Error() != "NO_PASSC" {
		return opt.Empty[Encryptor]()
	}

	tempEncryptor := Encryptor{
		FilePath:       tempFilePath,
		MasterPassword: strings.TrimSuffix(strings.TrimPrefix(tempFilePath, "/tmp/"), passcExtension),
	}
	password, err := tempEncryptor.ReadEncryptedText()
	if err != nil {
		return opt.Empty[Encryptor]()
	}

	return opt.Of(Encryptor{MasterPassword: password, FilePath: filePath})
}
