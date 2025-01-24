package passc

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Builder() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "passc",
	}
	rootCmd.AddCommand(version())
	rootCmd.AddCommand(list())

	return rootCmd
}

func version() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "passcualito version",
		Long:  "passcualito version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Passcualito v0.1.0")
		},
	}
}

func list() *cobra.Command {
	return &cobra.Command{
		Use:   "list [key]",
		Short: "List all properties of a key",
		Long:  "List all properties of a key. Ex: passc list my_key",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := createPasscualitoFileIfNotExist(); err != nil {
				fmt.Println(err.Error())
				return
			}

			if len(args) == 1 {
				fmt.Println("one")
			} else {
				fmt.Println("zero")
			}
		},
	}
}
