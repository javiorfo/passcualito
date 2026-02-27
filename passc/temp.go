package passc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/javiorfo/nilo"
)

func newTemp(masterPassword string) error {
	key, err := generateRandomPassword(nilo.Value(16), nilo.Nil[string]())
	filePath := fmt.Sprintf("%s/%s%s", os.TempDir(), *key, passcExtension)
	if err != nil {
		return fmt.Errorf("Error generating tmp password: %v", err)
	}
	encryptor := &Encryptor{MasterPassword: *key, FilePath: filePath}
	encryptor.encryptText(masterPassword, false)
	return nil
}

func removeTemp() error {
	err := filepath.Walk(os.TempDir(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == passcExtension {
			err := os.Remove(path)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("Error deleting from Temp directory %v", err)
	}
	return nil
}

func getTempEncryptor(filePath string) nilo.Option[Encryptor] {
	var tempFilePath string

	tempDir := os.TempDir()
	err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !info.IsDir() && filepath.Ext(path) == passcExtension {
			tempFilePath = path
		}

		return nil
	})

	if err != nil && tempFilePath == "" {
		return nilo.Nil[Encryptor]()
	}

	tempEncryptor := Encryptor{
		FilePath:       tempFilePath,
		MasterPassword: strings.TrimSuffix(strings.TrimPrefix(tempFilePath, tempDir+"/"), passcExtension),
	}
	password, err := tempEncryptor.readEncryptedText()
	if err != nil {
		return nilo.Nil[Encryptor]()
	}

	return nilo.Value(Encryptor{MasterPassword: password, FilePath: filePath})
}
