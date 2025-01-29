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
		Use: passcAppCommandName,
	}

	rootCmd.AddCommand(add()) // bkp
	rootCmd.AddCommand(copy())
	rootCmd.AddCommand(export())
	rootCmd.AddCommand(importer()) // bkp
	rootCmd.AddCommand(list())
	rootCmd.AddCommand(logout())
	rootCmd.AddCommand(password())
	rootCmd.AddCommand(remove()) // bkp
	rootCmd.AddCommand(version())

	return rootCmd
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
				if !errors.Is(err, emptyFile) {
					fmt.Println(passcInvalidPassword)
					removeTemp()
					return
				}
			}

			if isNameTaken(content, name) {
				fmt.Printf(passcNameTakenText, name)
				return
			}

			if err := encryptor.encryptText(*json, true); err != nil {
				log.Println("error encrypting data: ", err.Error())
				return
			}

			fmt.Printf(passcEntryCreatedText, name)
		},
	}

	add.Flags().StringVarP(&password, "password", "p", "", "Password for the entry")
	add.Flags().StringVarP(&info, "info", "i", "", "Additional info for the entry")

	return add
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

func importer() *cobra.Command {
	return &cobra.Command{
		Use:   "import [key]",
		Short: "Import entries from a JSON file",
		Long:  "Import entries from a JSON file. Ex: passc import /path/to/file.json",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			encryptor, err := checkMasterPassword()
			if err != nil {
				log.Println("checking Master Password: ", err.Error())
				return
			}

			filePath := args[0]
			dataFromJson, err := getDataSliceFromJsonFile(filePath)
			if err != nil {
				fmt.Println("Error getting data from JSON file:", err.Error())
				return
			}

			content, err := encryptor.readEncryptedText()
			if err != nil {
				if errors.Is(err, emptyFile) {
					if err := encryptDataSlice(encryptor, dataFromJson); err != nil {
						log.Println("encrypting data: ", err.Error())
					}
					fmt.Printf(passcImportText, filePath, len(dataFromJson))
					return
				} else {
					fmt.Println(passcInvalidPassword)
					removeTemp()
					return
				}
			}

			dataSlice := stringToDataSlice(content)
            // Filters matched names between file.json and the actual store
			repeatedSlice := steams.OfSlice(dataSlice).
				Filter(func(outer Data) bool {
					return steams.OfSlice(dataFromJson).
						AnyMatch(predicateMatchOuterData(outer))
				}).
				MapToString(func(d Data) string { return d.Name }).Collect()

			if len(repeatedSlice) != 0 {
				fmt.Printf(passcImportRepeatdText, filePath, strings.Join(repeatedSlice, ", "))
				return
			}

			if err := encryptor.deleteContent(); err != nil {
				log.Println("deleting file: ", err.Error())
				return
			}

			if err := encryptDataSlice(encryptor, dataSlice); err != nil {
				log.Println("encrypting data: ", err.Error())
			}
			if err := encryptDataSlice(encryptor, dataFromJson); err != nil {
				log.Println("encrypting data: ", err.Error())
			}
			fmt.Printf(passcImportText, filePath, len(dataFromJson))
		},
	}
}

func predicateMatchOuterData(outer Data) func(Data) bool {
	return func(inner Data) bool {
		return outer.Name == inner.Name
	}
}

func encryptDataSlice(encryptor *Encryptor, dataSlice []Data) error {
	for _, v := range dataSlice {
		json, err := v.toJSON()
		if err != nil {
			return fmt.Errorf("converting data to JSON: %v", err.Error())
		}
		if err := encryptor.encryptText(*json, true); err != nil {
			return fmt.Errorf("encrypting data: %v", err.Error())
		}
	}
	return nil
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
            var found bool
			for i, data := range steams.OfSlice(items).Sorted(sortByName).Collect() {
				isEnd := i == length-1
				if isSearcOne {
					name := args[0]
					if strings.Contains(strings.ToLower(data.Name),strings.ToLower(name)) {
						data.print(true)
                        found = true
					}
					if isEnd && !found {
						fmt.Printf(passcNameNotFoundText, name)
					}
				} else {
					data.print(isEnd)
				}
			}
		},
	}
}

func sortByName(d1, d2 Data) bool {
	return d1.Name < d2.Name
}

func logout() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Logout of the app",
		Long:  "Logout of the app. This allows you to enter the master password again",
		Run: func(cmd *cobra.Command, args []string) {
			err := removeTemp()
			_ = err // Ignored
			fmt.Println(passcLogoutText)
		},
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
				if errors.Is(err, emptyFile) {
					fmt.Println(err.Error())
					return
				}
				fmt.Println(passcInvalidPassword)
				removeTemp()
				return
			}

			contentLength := len(content)
			name := args[0]

			result := steams.OfSlice(strings.Split(content, passcItemSeparator)).
				Filter(predicateByName(name)).Reduce("", reducer)

			resultLength := len(result)
			// If lengths are equal, no entry has been cut off
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
				fmt.Printf(passcEntryRemovedText, name)
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
