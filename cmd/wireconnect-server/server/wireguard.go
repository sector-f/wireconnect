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
	for _, link := range s.activeInterfaces {
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

	s.activeInterfaces = append(s.activeInterfaces, link)

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
	if err != nil {
		return err
	}

	err = netlink.LinkSetUp(link)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) addPeer(username string, request wireconnect.ConnectionRequest) error {
	peerConfig := s.db.GetPeer(username, request.PeerName)
	if peerConfig == nil {
		return errors.New("Peer does not exist")
	}

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

	err = s.wgClient.ConfigureDevice(peerConfig.DBIface.Name, config)
	if err != nil {
		return err
	}

	usermap, present := s.activePeers[username]
	if !present {
		usermap = make(map[string]wgtypes.Key)
		s.activePeers[username] = usermap
	}
	usermap[request.PeerName] = key
	return nil
}

func (s *Server) removePeer(username, peername string) error {
	pubkey, present := s.activePeers[username][peername]
	if !present {
		return errors.New("Peer is not active")
	}

	peerConfig := s.db.GetPeer(username, peername)
	if peerConfig == nil {
		return errors.New("Peer does not exist")
	}

	dev, err := s.wgClient.Device(peerConfig.DBIface.Name)
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
				PublicKey:  pubkey,
				Remove:     true,
				UpdateOnly: true,
			},
		},
	}

	err = s.wgClient.ConfigureDevice(peerConfig.DBIface.Name, config)
	if err != nil {
		return err
	}

	delete(s.activePeers[username], peername)
	return nil
}

func (s *Server) Shutdown() {
	log.Println("Shutting down")

	for _, link := range s.activeInterfaces {
		log.Printf("Deleting interface: %s\n", link.Attrs().Name)
		err := netlink.LinkDel(link)
		if err != nil {
			log.Println(err)
		}
	}
}
