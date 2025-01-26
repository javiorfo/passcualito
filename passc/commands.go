package passc

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func Builder() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "passc",
	}
	rootCmd.AddCommand(version())
	rootCmd.AddCommand(logout())
	rootCmd.AddCommand(list())
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
			data := Data{
				Name:     name,
				Password: password,
				Info:     info,
			}
			json, err := data.ToJSON()
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

			if len(args) == 1 {
				fmt.Println("one")
			} else {
				content, err := encryptor.ReadEncryptedText()
				if err != nil {
					// TODO logger
					fmt.Println(passcInvalidPassword)
					removeTemp()
					return
				}
				fmt.Println(passcStoreTitle)
				items := strings.Split(content, passcItemSeparator)
				length := len(items)
				for i, v := range items {
					var data Data
					err = data.FromJSON([]byte(v))
					if err != nil {
						// TODO logger
						fmt.Println(err.Error())
						return
					}
					fmt.Println("│")
					fmt.Println("├─ \033[1mName:\033[0m", data.Name)
					fmt.Println("├─── \033[1mPassword:\033[0m", data.Password)
					if i == length-1 {
						fmt.Println("└─── \033[1mInfo:\033[0m", data.Info)
					} else {
						fmt.Println("├─── \033[1mInfo:\033[0m", data.Info)

					}
				}

			}
		},
	}
}
