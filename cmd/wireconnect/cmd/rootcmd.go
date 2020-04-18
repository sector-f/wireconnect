package cmd

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	Username string
	Password string
	Server   string
	Client   *http.Client
)

func Root() *cobra.Command {
	rootCmd := cobra.Command{
		Use:           "wireconnect -u USERNAME[:PASSWORD] -s SERVER[:IP] SUBCOMMAND [flags]",
		Short:         "Client for the wireconnect VPN server",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			serverarg, err := cmd.Flags().GetString("server")
			if err != nil {
				return err
			}

			if serverarg == "" {
				return errors.New("Server not specified")
			}

			userarg, err := cmd.Flags().GetString("user")
			if err != nil {
				return err
			}

			if userarg == "" {
				return errors.New("Username not specified")
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

			Username = username
			Password = password

			insecurearg, err := cmd.Flags().GetBool("insecure")
			if err != nil {
				return err
			}

			config := &tls.Config{
				InsecureSkipVerify: insecurearg,
			}

			transport := &http.Transport{
				TLSClientConfig: config,
			}

			Client = &http.Client{
				Transport: transport,
			}

			return nil
		},
	}

	rootCmd.PersistentFlags().StringP("user", "u", "", "Specify username[:password]")
	rootCmd.PersistentFlags().StringP("server", "s", "", "Specify server address[:port] (Default port: 8900)")
	rootCmd.PersistentFlags().BoolP("insecure", "k", false, "Ignore insecure TLS connections")
	rootCmd.AddCommand(connectCmd())

	return &rootCmd
}
