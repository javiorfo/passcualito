package passc

import (
	"errors"
	"fmt"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"os"
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
		finalText += passcItemSeparator + decryptedText
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
		return "", fmt.Errorf("Error open file %s: %v", e.FilePath, err)
	}
	defer file.Close()

	block, err := aes.NewCipher([]byte(e.MasterPassword))
	if err != nil {
		return "", fmt.Errorf("Error cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("Error gcm: %v", err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("Error reading file: %v", err)
	}

	nonce := data[:gcm.NonceSize()]
	if len(data) == 0 {
		return "", errors.New(passcEmptyFile)
	}
	ciphertext := data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("Error extracting plaintext: %v", err)
	}

	return string(plaintext), nil
}
