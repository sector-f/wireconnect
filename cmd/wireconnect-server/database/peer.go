package database

import (
	"github.com/sector-f/wireconnect"
)

type PeerConfig struct {
	Name    string
	Address wireconnect.Address
	DBIface *DBIface
}

func (s *ServiceDB) CreatePeer(peer wireconnect.CreatePeerRequest) error {
	peerAddr, err := wireconnect.ParseAddress(peer.Address)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		`INSERT INTO peers (name, address, mask, server_interface_id, user_id)
		VALUES (
			?,
			?,
			?,
			(SELECT id FROM server_interfaces WHERE name = ?),
			(SELECT id FROM users WHERE username = ?)
		)`,
		peer.PeerName,
		peerAddr.Address,
		peerAddr.Mask,
		peer.ServerInterface,
		peer.UserName,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *ServiceDB) GetPeer(username, peername string) *PeerConfig {
	row := s.db.QueryRow(
		`SELECT address, mask, server_interface_id
		FROM peers
		WHERE user_id = (SELECT id FROM users WHERE username = ?)
		AND name = ?`,
		username,
		peername,
	)

	var addr wireconnect.Address
	var ifaceID int

	err := row.Scan(&addr.Address, &addr.Mask, &ifaceID)
	if err != nil {
		return nil
	}

	iface, err := s.getIfaceFromID(ifaceID)
	if err != nil {
		return nil
	}

	return &PeerConfig{
		Name:    peername,
		Address: addr,
		DBIface: iface,
	}
}

func (s *ServiceDB) ListPeers(username string) *[]wireconnect.Peer {
	rows, err := s.db.Query(
		`SELECT name FROM peers INNER JOIN users ON users.id = peers.user_id WHERE users.username = ?`,
		username,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	peerNames := []string{}
	for rows.Next() {
		var peerName string
		if err := rows.Scan(&peerName); err != nil {
			return nil
		}

		peerNames = append(peerNames, peerName)
	}

	peerConfigs := []PeerConfig{}
	for _, peerName := range peerNames {
		config := s.GetPeer(username, peerName)
		if config != nil {
			peerConfigs = append(peerConfigs, *config)
		}
	}
	if len(peerConfigs) == 0 {
		return nil
	}

	peers := []wireconnect.Peer{}
	for _, config := range peerConfigs {
		peers = append(
			peers,
			wireconnect.Peer{
				Name:            config.Name,
				Address:         config.Address.String(),
				ServerInterface: config.DBIface.Name,
			},
		)
	}

	return &peers
}
