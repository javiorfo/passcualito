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

	rootCmd.AddCommand(add())
	rootCmd.AddCommand(copy())
	rootCmd.AddCommand(edit())
	rootCmd.AddCommand(export())
	rootCmd.AddCommand(importer())
	rootCmd.AddCommand(list())
	rootCmd.AddCommand(logout())
	rootCmd.AddCommand(password())
	rootCmd.AddCommand(remove())
	rootCmd.AddCommand(version())

	return rootCmd
}

func add() *cobra.Command {
	var password string
	var info string

	add := &cobra.Command{
		Use:     "add [name]",
		Short:   "Add a new entry to the store",
		Long:    "Add a new entry to the store. Password (-p flag) and Info (-i flag) are optionals",
		Example: "passc add acme -p p4$$w0rd -i \"acme.com page\"",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			encryptor, err := checkMasterPassword()
			if err != nil {
				if errors.Is(err, masterPasswordError) {
					fmt.Println(err)
					return
				}
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
			data.print(true)
			makeBackUp()
		},
	}

	add.Flags().StringVarP(&password, "password", "p", "", "Password for the entry")
	add.Flags().StringVarP(&info, "info", "i", "", "Additional info for the entry")

	return add
}

func copy() *cobra.Command {
	return &cobra.Command{
		Use:     "copy [name]",
		Short:   "Copy password to clipboard",
		Long:    "Copy password to clipboard",
		Example: "passc copy name_here",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			encryptor, err := checkMasterPassword()
			if err != nil {
				if errors.Is(err, masterPasswordError) {
					fmt.Println(err)
					return
				}
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

func edit() *cobra.Command {
	var password string
	var info string
	edit := &cobra.Command{
		Use:     "edit [name]",
		Short:   "Edit the entry.",
		Long:    "Edit the entry. Password with -p and Info with -i. Each one is optional",
		Example: "passc edit name_here -p 1234 -i \"some info\"",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			encryptor, err := checkMasterPassword()
			if err != nil {
				if errors.Is(err, masterPasswordError) {
					fmt.Println(err)
					return
				}
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

			// result of all data without the input name
			var data Data
			result := steams.OfSlice(strings.Split(content, passcItemSeparator)).
				Filter(predicateByNameAndSetData(name, &data)).Reduce("", reducer)

			resultLength := len(result)
			// If lengths are equal, no entry has been cut off
			if contentLength == resultLength {
				fmt.Printf(passcNameNotFoundText, name)
			} else {
				if password != "" {
					data.Password = password
				}
				if info != "" {
					data.Info = info
				}
				json, err := data.toJSON()
				if err != nil {
					log.Println("converting data to JSON: ", err.Error())
					return
				}

				// Add edited item to final string
				result = result + passcItemSeparator + *json

				// override store.passc completly
				if err := encryptor.encryptText(result, false); err != nil {
					log.Println("encryting text: ", err.Error())
					return
				}
				fmt.Printf(passcEntryEditedText, name)
				data.print(true)
				makeBackUp()
			}
		},
	}

	edit.Flags().StringVarP(&password, "password", "p", "", "Password for the entry")
	edit.Flags().StringVarP(&info, "info", "i", "", "Additional info for the entry")

	return edit
}

func predicateByNameAndSetData(name string, data *Data) func(string) bool {
	return func(value string) bool {
		isInput := strings.Contains(value, fmt.Sprintf(`"name":"%s"`, name))
		if isInput {
			data.fromJSON([]byte(value))
		}
		return !isInput
	}
}

func export() *cobra.Command {
	return &cobra.Command{
		Use:     "export",
		Short:   "Export data in a JSON file",
		Long:    "Export data in a JSON file",
		Example: "passc export",
		Run: func(cmd *cobra.Command, args []string) {
			encryptor, err := checkMasterPassword()
			if err != nil {
				if errors.Is(err, masterPasswordError) {
					fmt.Println(err)
					return
				}
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
		Use:     "import [filepath]",
		Short:   "Import entries from a JSON file",
		Long:    "Import entries from a JSON file",
		Example: "passc import /path/to/file.json",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			encryptor, err := checkMasterPassword()
			if err != nil {
				if errors.Is(err, masterPasswordError) {
					fmt.Println(err)
					return
				}
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
			makeBackUp()
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
		Use:     "list [name]",
		Short:   "List all properties of the entry by name",
		Long:    "List all properties of the entry by name",
		Example: "passc list name_here",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			encryptor, err := checkMasterPassword()
			if err != nil {
				if errors.Is(err, masterPasswordError) {
					fmt.Println(err)
					return
				}
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
					if strings.Contains(strings.ToLower(data.Name), strings.ToLower(name)) {
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
		Use:     "logout",
		Short:   "Logout of the app",
		Long:    "Logout of the app. This allows you to enter the master password again",
		Example: "passc logout",
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
		Use:     "password [number]",
		Short:   "Generates a password of the number passed",
		Long:    "Generates a password of the number passed. Charset (-c flag) could be a, n, an or anc",
		Example: "passc password 12 -c an",
		Args:    cobra.ExactArgs(1),
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
		Use:     "remove [name]",
		Short:   "Remove the entry",
		Long:    "Remove the entry if exists",
		Example: "passc remove entry_name",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			encryptor, err := checkMasterPassword()
			if err != nil {
				if errors.Is(err, masterPasswordError) {
					fmt.Println(err)
					return
				}
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

			// result of all data without the input name
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
				makeBackUp()
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
		Use:     "version",
		Short:   "app version",
		Long:    "app version",
		Example: "passc version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(passcVersion)
		},
	}
}
