package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"

	"github.com/sector-f/wireconnect"
	"github.com/spf13/cobra"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
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

			privKey, err := wgtypes.GeneratePrivateKey()
			if err != nil {
				return err
			}

			pubKey := privKey.PublicKey()

			msg := wireconnect.ConnectionRequest{
				PeerName:  args[0],
				PublicKey: pubKey.String(),
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
					Path:   "/connect",
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

			if resp.StatusCode == http.StatusOK {
				data, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return err
				}

				var reply wireconnect.ConnectionReply
				err = json.Unmarshal(data, &reply)
				if err != nil {
					return err
				}

				addr, network, err := net.ParseCIDR(reply.ClientAddress)
				if err != nil {
					return err
				}

				endpointAddr := net.ParseIP(reply.EndpointAddress)
				if endpointAddr == nil {
					return errors.New("Invalid endpoint address")
				}

				// FIXME: should creation of the WireGuard interface be
				// before or after connecting to the server?

				wgClient, err := wgctrl.New()
				if err != nil {
					return err
				}

				serverPubKey, err := wgtypes.ParseKey(reply.PublicKey)
				if err != nil {
					return err
				}

				linkAttrs := netlink.NewLinkAttrs()
				linkAttrs.Name = "wireconnect"
				link := &netlink.GenericLink{
					linkAttrs,
					"wireguard",
				}

				err = netlink.LinkAdd(link)
				if err != nil {
					return err
				}

				netAddr := &net.IPNet{
					IP:   addr,
					Mask: network.Mask,
				}

				nlAddr := netlink.Addr{IPNet: netAddr}

				err = netlink.AddrAdd(link, &nlAddr)
				if err != nil {
					return err
				}

				wgConfig := wgtypes.Config{
					PrivateKey: &privKey,
					Peers: []wgtypes.PeerConfig{
						wgtypes.PeerConfig{
							PublicKey:    serverPubKey,
							Remove:       false,
							UpdateOnly:   false,
							PresharedKey: nil,
							Endpoint: &net.UDPAddr{
								IP:   endpointAddr,
								Port: reply.EndpointPort,
							},
							PersistentKeepaliveInterval: nil,
							ReplaceAllowedIPs:           true, // Probably not needed
							AllowedIPs: []net.IPNet{
								net.IPNet{
									IP:   network.IP,
									Mask: network.Mask,
								},
							},
						},
					},
				}

				err = wgClient.ConfigureDevice("wireconnect", wgConfig)
				if err != nil {
					return err
				}

				err = netlink.LinkSetUp(link)
				if err != nil {
					return err
				}
			} else {

				data, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return err
				}

				var reply string
				err = json.Unmarshal(data, reply)
				if err != nil {
					return err
				}

				fmt.Printf("Received %v: %v\n", resp.Status, reply)
			}

			return nil
		},
	}

	return &connectCmd
}
