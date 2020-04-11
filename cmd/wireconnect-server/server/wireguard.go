package server

import (
	"errors"
	"log"
	"net"

	"github.com/sector-f/wireconnect"
	"github.com/sector-f/wireconnect/cmd/wireconnect-server/database"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func (s *Server) makeIface(iface *database.DBIface) error {
	for _, link := range s.active {
		if link.Attrs().Name == iface.Name {
			if link.Type() == "wireguard" {
				return nil
			} else {
				return errors.New("Interface exists but is not WireGuard interface")
			}
		}
	}

	linkAttrs := netlink.NewLinkAttrs()

	linkAttrs.Name = iface.Name
	link := &netlink.GenericLink{
		linkAttrs,
		"wireguard",
	}

	err := netlink.LinkAdd(link)
	if err != nil {
		return err
	}

	s.active = append(s.active, link)

	for _, addr := range iface.Addresses {
		log.Printf("\t%v/%v\n", addr.Address, cidr(addr.Mask))

		netAddr := &net.IPNet{
			IP:   addr.Address,
			Mask: addr.Mask,
		}

		nlAddr := netlink.Addr{IPNet: netAddr}

		err = netlink.AddrAdd(link, &nlAddr)
		if err != nil {
			return err
		}
	}

	privkey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return err
	}

	wgConfig := wgtypes.Config{
		PrivateKey: &privkey,
	}

	err = s.wgClient.ConfigureDevice(iface.Name, wgConfig)

	return nil
}

func (s *Server) addPeer(request wireconnect.ConnectionRequest, peerConfig *database.PeerConfig) error {
	dev, err := s.wgClient.Device(peerConfig.DBIface.Name)
	if err != nil {
		return err
	}

	key, err := wgtypes.ParseKey(request.PublicKey)
	if err != nil {
		return err
	}

	config := wgtypes.Config{
		PrivateKey:   &dev.PrivateKey,
		ListenPort:   &dev.ListenPort,
		FirewallMark: &dev.FirewallMark,
		ReplacePeers: false,
		Peers: []wgtypes.PeerConfig{
			wgtypes.PeerConfig{
				PublicKey:                   key,
				Remove:                      false,
				UpdateOnly:                  false,
				PresharedKey:                nil,
				Endpoint:                    nil, // TODO: Get this from http request IP?
				PersistentKeepaliveInterval: nil,
				ReplaceAllowedIPs:           true, // Probably not needed
				AllowedIPs: []net.IPNet{
					net.IPNet{
						IP:   peerConfig.Address.Address,
						Mask: net.IPv4Mask(255, 255, 255, 255),
					},
				},
			},
		},
	}

	return s.wgClient.ConfigureDevice(peerConfig.DBIface.Name, config)
}

func (s *Server) shutdown() {
	log.Println("Shutting down")

	for _, link := range s.active {
		log.Printf("Deleting interface: %s\n", link.Attrs().Name)
		err := netlink.LinkDel(link)
		if err != nil {
			log.Println(err)
		}
	}
}
