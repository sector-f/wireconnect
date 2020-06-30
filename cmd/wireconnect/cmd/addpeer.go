package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/sector-f/wireconnect"
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
				fmt.Print("Enter peer configuration name: ")
				name, _ = reader.ReadString('\n')
			}

			username, _ := cmd.Flags().GetString("username")
			if username == "" {
				fmt.Print("Enter username: ")
				username, _ = reader.ReadString('\n')
			}

			address, _ := cmd.Flags().GetString("address")
			if address == "" {
				fmt.Print("Enter peer address: ")
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

			msg := &wireconnect.CreatePeerRequest{
				UserName:        username,
				PeerName:        name,
				Address:         address,
				EndpointAddress: endpointAddress,
				ServerInterface: serverInterface,
			}

			jsonMsg, err := json.Marshal(msg)
			if err != nil {
				return err
			}

			req := &http.Request{
				Method: "POST",
				URL: &url.URL{
					Scheme: "https",
					Host:   Server,
					Path:   "/peers",
				},
				Body:   ioutil.NopCloser(bytes.NewBuffer(jsonMsg)),
				Header: make(http.Header),
			}

			req.Header.Add("Content-Type", "application/json")
			req.SetBasicAuth(Username, Password)

			resp, err := Client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusCreated {
				_, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return err
				}

				fmt.Println("Peer created")
			} else {
				data, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return err
				}

				var reply string
				err = json.Unmarshal(data, &reply)
				if err != nil {
					return err
				}

				fmt.Printf("Received %v: %v\n", resp.Status, reply)
			}

			return nil
		},
	}

	addPeerCmd.Flags().StringP("name", "n", "", "Peer configuration name")
	addPeerCmd.Flags().String("username", "", "Username of peer configuration's owner") // FIXME: can't use -u here because it's used in rootCmd
	addPeerCmd.Flags().StringP("address", "a", "", "Peer's WireGuard address")
	addPeerCmd.Flags().StringP("endpoint-address", "e", "", "Endpoint address for peer to connect to")
	addPeerCmd.Flags().StringP("server-interface", "i", "", "WireGuard interface on the server")

	return &addPeerCmd
}
