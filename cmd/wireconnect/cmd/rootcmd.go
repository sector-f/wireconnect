package cmd

import (
	"errors"
	"fmt"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

func Root() *cobra.Command {
	rootCmd := cobra.Command{
		Use:           "wireconnect [-u USERNAME[:PASSWORD]] [-s SERVER[:IP]] SUBCOMMAND [flags]",
		Short:         "Client for the wireconnect VPN server",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			userarg, err := cmd.Flags().GetString("user")
			if err != nil {
				return err
			}

			if userarg == "" {
				return errors.New("No username specified")
			}

			userpass := strings.SplitN(userarg, ":", 2)
			var password string

			if len(userpass) == 1 {
				fmt.Print("Password: ")
				pw, err := terminal.ReadPassword(int(syscall.Stdin))
				if err != nil {
					return err
				}
				fmt.Println()
				password = string(pw)
			} else {
				password = userpass[1]
			}

			username := userpass[0]

			fmt.Printf("Username: \"%s\"\n", username)
			fmt.Printf("Password: \"%s\"\n", password)

			return nil
		},
	}

	rootCmd.PersistentFlags().StringP("user", "u", "", "Specify username[:password]")

	return &rootCmd
}
