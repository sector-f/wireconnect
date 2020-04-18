package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

func connectCmd() *cobra.Command {
	connectCmd := cobra.Command{
		Use:           "connect PEERNAME",
		Short:         "Connect to wireconnect VPN server",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("No peer specified")
			} else if len(args) > 1 {
				return errors.New("Too many arguments specified")
			}

			return errors.New("Unimplemented subcommand")
		},
	}

	return &connectCmd
}
