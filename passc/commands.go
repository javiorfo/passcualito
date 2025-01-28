package passc

import (
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

func Builder() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "passc",
	}
	rootCmd.AddCommand(version())
	rootCmd.AddCommand(logout())
	rootCmd.AddCommand(list())
	rootCmd.AddCommand(copy())
	rootCmd.AddCommand(add())

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

func logout() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Logout of the app",
		Long:  "Logout of the app. This allows you to enter the master password again",
		Run: func(cmd *cobra.Command, args []string) {
			removeTemp()
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
				// TODO logger
				fmt.Println(err.Error())
				return
			}
			name := args[0]

			// TODO check password
			data := Data{
				Name:     name,
				Password: password,
				Info:     info,
			}
			json, err := data.toJSON()
			if err != nil {
				// TODO logger
				fmt.Println(err.Error())
				return
			}
			// TODO check "name" does not exist
			if err := encryptor.EncryptText(*json); err != nil {
				// TODO logger
				fmt.Println(passcInvalidPassword)
				removeTemp()
				return
			}

			fmt.Println("Done")
		},
	}

	add.Flags().StringVarP(&password, "password", "p", "", "Password for the entry")
	add.Flags().StringVarP(&info, "info", "i", "", "Additional info for the entry")

	return add
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
				// TODO logger
				fmt.Println(err.Error())
				return
			}

			content, err := encryptor.ReadEncryptedText()
			if err != nil {
				// TODO logger
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
				// TODO logger
				fmt.Println(err.Error())
				return
			}

			content, err := encryptor.ReadEncryptedText()
			if err != nil {
				// TODO logger
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
						// TODO err clipboard
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
