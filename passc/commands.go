package passc

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/javiorfo/steams"
	"github.com/javiorfo/steams/opt"
	"github.com/spf13/cobra"
)

func Builder() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "passc",
	}
	rootCmd.AddCommand(add())
	rootCmd.AddCommand(copy())
	rootCmd.AddCommand(export())
	rootCmd.AddCommand(list())
	rootCmd.AddCommand(logout())
	rootCmd.AddCommand(password())
	rootCmd.AddCommand(remove())
	rootCmd.AddCommand(version())

	return rootCmd
}

func version() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "passcualito version",
		Long:  "passcualito version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(passcVersion)
		},
	}
}

func export() *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Export data in a JSON file",
		Long:  "Export data in a JSON file",
		Run: func(cmd *cobra.Command, args []string) {
			encryptor, err := checkMasterPassword()
			if err != nil {
				log.Println("checking Master Password: ", err.Error())
				return
			}

			content, err := encryptor.readEncryptedText()
			if err != nil {
				if errors.Is(err, emptyFile) {
					fmt.Println(err.Error())
					return
				}
				fmt.Println(passcInvalidPassword)
				removeTemp()
				return
			}

			if err := exportToFile(content); err != nil {
				fmt.Println(passcExportErr)
			} else {
				fmt.Printf(passcExportText, passcExportFilename)
			}
		},
	}
}

func logout() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Logout of the app",
		Long:  "Logout of the app. This allows you to enter the master password again",
		Run: func(cmd *cobra.Command, args []string) {
			err := removeTemp()
			if err != nil {
				log.Println("removing temp file: ", err.Error())
				return
			}
			fmt.Println(passcLogoutText)
		},
	}
}

func add() *cobra.Command {
	var password string
	var info string

	add := &cobra.Command{
		Use:   "add [name]",
		Short: "Add a new entry to protect",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			encryptor, err := checkMasterPassword()
			if err != nil {
				log.Println("checking Master Password: ", err.Error())
				return
			}
			name := args[0]

			data, err := newData(name, password, info)
			if err != nil {
				log.Println("generating random password: ", err.Error())
				return
			}

			json, err := data.toJSON()
			if err != nil {
				log.Println("converting data to JSON: ", err.Error())
				return
			}

			content, err := encryptor.readEncryptedText()
			if err != nil {
				fmt.Println(passcInvalidPassword)
				removeTemp()
				return
			}

			if isNameTaken(content, name) {
				fmt.Printf(passcNameTakenText, name)
				return
			}

			if err := encryptor.encryptText(*json, true); err != nil {
				return
			}

			fmt.Printf(passcEntryCreatedText, name)
		},
	}

	add.Flags().StringVarP(&password, "password", "p", "", "Password for the entry")
	add.Flags().StringVarP(&info, "info", "i", "", "Additional info for the entry")

	return add
}

func remove() *cobra.Command {
	return &cobra.Command{
		Use:   "remove [name]",
		Short: "Remove the entry",
		Long:  "Remove the entry. Ex: passc remove entry_name",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			encryptor, err := checkMasterPassword()
			if err != nil {
				log.Println("checking Master Password: ", err.Error())
				return
			}

			content, err := encryptor.readEncryptedText()
			if err != nil {
				fmt.Println(passcInvalidPassword)
				removeTemp()
				return
			}

			contentLength := len(content)
			name := args[0]

			result := steams.OfSlice(strings.Split(content, passcItemSeparator)).
				Filter(predicateByName(name)).Reduce("", reducer)

			resultLength := len(result)
			if contentLength == resultLength {
				fmt.Printf(passcNameNotFoundText, name)
			} else {
				if resultLength == 0 {
					if err := encryptor.deleteContent(); err != nil {
						log.Println("deleting file: ", err.Error())
						return
					}
				} else {
					if err := encryptor.encryptText(result, false); err != nil {
						log.Println("encryting text: ", err.Error())
						return
					}
				}
			}
		},
	}
}

func reducer(a, b string) string {
	if a == "" {
		return b
	} else {
		return a + passcItemSeparator + b
	}
}

func predicateByName(name string) func(string) bool {
	return func(value string) bool {
		return !strings.Contains(value, fmt.Sprintf(`"name":"%s"`, name))
	}
}

func password() *cobra.Command {
	var charset string
	password := &cobra.Command{
		Use:   "password [number]",
		Short: "Generates a password of the number passed",
		Long:  "Generates a password of the number passed. Ex: passc password 12",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			number, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Println(passcPasswordParamErrNumber)
				return
			}
			if number < 1 {
				fmt.Println(passcPasswordParamErrPositive)
				return
			}

			var optCharset opt.Optional[string]
			switch charset {
			case "n":
				optCharset = opt.Of(passcCharsetNumeric)
			case "a":
				optCharset = opt.Of(passcCharsetAlpha)
			case "an":
				optCharset = opt.Of(passcCharsetAlphaNumeric)
			case "anc":
				optCharset = opt.Of(passcCharsetAlphaNumericCap)
			default:
				optCharset = opt.Of(passcCharset)
			}

			psswd, err := generateRandomPassword(opt.Of(number), optCharset)
			if err != nil {
				log.Println("generating random password: ", err.Error())
				return
			}
			fmt.Println(*psswd)
		},
	}
	password.Flags().StringVarP(&charset, "charset", "c", "", "Charset for the password (a, n, an or anc)")

	return password
}

func list() *cobra.Command {
	return &cobra.Command{
		Use:   "list [key]",
		Short: "List all properties of a key",
		Long:  "List all properties of a key. Ex: passc list my_key",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			encryptor, err := checkMasterPassword()
			if err != nil {
				log.Println("checking Master Password: ", err.Error())
				return
			}

			content, err := encryptor.readEncryptedText()
			if err != nil {
				if errors.Is(err, emptyFile) {
					fmt.Println(err.Error())
					return
				}
				fmt.Println(passcInvalidPassword)
				removeTemp()
				return
			}

			items := stringToDataSlice(content)
			length := len(items)
			if length == 0 {
				fmt.Println(passcEmptyFile)
				return
			}

			isSearcOne := len(args) == 1

			fmt.Println(passcStoreTitle)
			for i, data := range items {
				isEnd := i == length-1
				if isSearcOne {
					name := args[0]
					if data.Name == name {
						data.print(true)
						break
					}
					if isEnd {
						fmt.Printf(passcNameNotFoundText, name)
					}
				} else {
					data.print(isEnd)
				}
			}
		},
	}
}

func copy() *cobra.Command {
	return &cobra.Command{
		Use:   "copy [key]",
		Short: "Copy password to clipboard",
		Long:  "Copy password to clipboard. Ex: passc copy my_key",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			encryptor, err := checkMasterPassword()
			if err != nil {
				log.Println("checking Master Password: ", err.Error())
				return
			}

			content, err := encryptor.readEncryptedText()
			if err != nil {
				fmt.Println(passcInvalidPassword)
				removeTemp()
				return
			}
			items := stringToDataSlice(content)
			length := len(items)
			for i, data := range items {
				isEnd := i == length-1
				name := args[0]
				if data.Name == name {
					err = clipboard.WriteAll(data.Password)
					if err != nil {
						log.Println("copying to clipboard: ", err.Error())
					} else {
						fmt.Printf(passcClipboardText, name)
					}
					break
				}
				if isEnd {
					fmt.Printf(passcNameNotFoundText, name)
				}

			}
		},
	}
}
