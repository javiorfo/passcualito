package passc

import (
	"fmt"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"os"
	"path/filepath"

	"golang.org/x/term"
)

var filePath string

func createPasscualitoFileIfNotExist() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("Error finding user home dir: %v", err)
	}

	dirPath := filepath.Join(homeDir, ".passcualito")
	filePath = filepath.Join(dirPath, "keys.passc")

	var masterPassword string
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Println("\033[1mPascualito Initialization\033[0m ")
		fmt.Print("Create Master Password: ")
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("Error reading password: %v", err)
		}
		masterPassword = alignPassword(string(bytePassword))

		err = os.MkdirAll(dirPath, 0755)
		if err != nil {
			return fmt.Errorf("Error creating directory: %v", err)
		}

		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("Error creating file: %v", err)
		}
		defer file.Close()
	} else {
		// TODO unlock file
	}
	fmt.Println("\nPassword entered:", masterPassword)
	return nil
}

func AppendEncryptedText(text, key string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	finalText := text
	stat, _ := file.Stat()
	if stat.Size() != 0 {
		decryptedText, err := ReadEncryptedText(key)
		if err != nil {
			fmt.Println("guacho:", err)
			return err
		}
		finalText += ";" + decryptedText
	}

	block, err := aes.NewCipher([]byte(key))
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

func ReadEncryptedText(key string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	block, err := aes.NewCipher([]byte(key))
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
	ciphertext := data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func main2() {
	key := "my_secret_key123"
	text := "name=github,pass=34h1oh4o1,username=,web="

	if err := AppendEncryptedText(text, key); err != nil {
		fmt.Println("Error appending encrypted text:", err)
		return
	}
	fmt.Println("Encrypted text appended to the file.")

	decryptedText, err := ReadEncryptedText(key)
	if err != nil {
		fmt.Println("Error reading encrypted text:", err)
		return
	}
	fmt.Println("Decrypted text:", decryptedText)
}
