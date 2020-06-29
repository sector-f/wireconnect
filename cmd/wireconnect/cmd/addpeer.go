package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func addPeerCmd() *cobra.Command {
	addPeerCmd := cobra.Command{
		Use:           "add-peer PEERNAME",
		Short:         "Create a peer configuration on the wireconnect VPN server",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(os.Stdin)

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				fmt.Print("Enter name: ")
				name, _ = reader.ReadString('\n')
			}

			username, _ := cmd.Flags().GetString("username")
			if username == "" {
				fmt.Print("Enter username: ")
				username, _ = reader.ReadString('\n')
			}

			address, _ := cmd.Flags().GetString("address")
			if address == "" {
				fmt.Print("Enter address: ")
				address, _ = reader.ReadString('\n')
			}

			endpointAddress, _ := cmd.Flags().GetString("endpoint-address")
			if endpointAddress == "" {
				fmt.Print("Enter endpoint address: ")
				endpointAddress, _ = reader.ReadString('\n')
			}

			serverInterface, _ := cmd.Flags().GetString("server-interface")
			if serverInterface == "" {
				fmt.Print("Enter server interface: ")
				serverInterface, _ = reader.ReadString('\n')
			}

			return errors.New("Unimplemented subcommand")
		},
	}

	addPeerCmd.Flags().StringP("name", "n", "", "Peer configuration name")
	addPeerCmd.Flags().String("username", "", "Username of peer configuration's owner") // FIXME: can't use -u here because it's used in rootCmd
	addPeerCmd.Flags().StringP("address", "a", "", "Peer's WireGuard address")
	addPeerCmd.Flags().StringP("endpoint-address", "e", "", "Endpoint address for peer to connect to")
	addPeerCmd.Flags().StringP("server-interface", "i", "", "WireGuard interface on the server")

	return &addPeerCmd
}
