package passc

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/howeyc/gopass"
	"github.com/javiorfo/steams/opt"
)

var masterPasswordError = errors.New(passcMasterPasswordLenErr)

func generateRandomPasswordDefault() (*string, error) {
	return generateRandomPassword(opt.Empty[int](), opt.Empty[string]())
}

func generateRandomPassword(size opt.Optional[int], charsetStr opt.Optional[string]) (*string, error) {
	length := size.OrElse(20)
	charset := charsetStr.OrElse(passcCharset)
	password := make([]byte, length)

	_, err := rand.Read(password)
	if err != nil {
		return nil, err
	}

	for i := 0; i < length; i++ {
		password[i] = charset[int(password[i])%len(charset)]
	}

	str := string(password)
	return &str, nil
}

func alignPassword(password string) string {
	length := len(password)
	if length < 16 {
		return password + strings.Repeat("*", 16-length)
	}
	if length > 16 {
		return password[:16]
	}
	return password
}

func checkMasterPassword() (*Encryptor, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("finding user home dir: %v", err)
	}

	dirPath := filepath.Join(homeDir, passcDirFolder)
	filePath := filepath.Join(dirPath, passcKeysFile)

	var masterPassword string
	_, errFile := os.Stat(filePath)

	if os.IsNotExist(errFile) {
		fmt.Println(passcInitTitle)
	} else {
		encryptorFromTemp := getTempEncryptor(filePath)
		if encryptorFromTemp.IsPresent() {
			e := encryptorFromTemp.Get()
			return &e, nil
		}
		fmt.Println(passcLoginTitle)
	}

	fmt.Print(passcMasterPasswordText)
	bytePassword, err := gopass.GetPasswdMasked()
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	if len(bytePassword) < 6 {
		return nil, masterPasswordError
	}
	masterPassword = alignPassword(string(bytePassword))

	if errFile != nil {
		err = os.MkdirAll(dirPath, 0755)
		if err != nil {
			return nil, fmt.Errorf("creating directory: %v", err)
		}

		file, err := os.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("creating file: %v", err)
		}
		defer file.Close()
	}

	newTemp(masterPassword)
	return &Encryptor{MasterPassword: masterPassword, FilePath: filePath}, nil
}
