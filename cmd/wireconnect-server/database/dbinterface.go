package database

import (
	"database/sql"
	"net"

	"github.com/sector-f/wireconnect"
)

type DBIface struct {
	Name            string
	CreateOnStartup bool
	Addresses       []wireconnect.Address
}

func (s *ServiceDB) Interface(name string) (*DBIface, error) {
	row := s.db.QueryRow(
		`SELECT id FROM server_interfaces WHERE name = ?`,
		name,
	)

	var id int

	err := row.Scan(&id)
	if err != nil {
		return nil, err
	}

	return s.getIfaceFromID(id)
}

func (s *ServiceDB) Ifaces() ([]DBIface, error) {
	rows, err := s.db.Query(`SELECT id FROM server_interfaces`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := []int{}

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	ifaces := []DBIface{}

	for _, id := range ids {
		iface, err := s.getIfaceFromID(id)
		if err != nil {
			return nil, err
		}
		ifaces = append(ifaces, *iface)
	}

	return ifaces, nil
}

func (s *ServiceDB) IfaceCount() (uint, error) {
	var count uint

	row := s.db.QueryRow(`SELECT COUNT(*) FROM server_interfaces`)
	switch err := row.Scan(&count); err {
	case sql.ErrNoRows, nil:
		return count, nil
	default:
		return 0, err
	}
}

func (s *ServiceDB) AddIface(iface DBIface) error {
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO server_interfaces (name) VALUES (?)`,
		iface.Name,
	)
	if err != nil {
		return err
	}

	var ifaceID int
	row := s.db.QueryRow(`SELECT id FROM server_interfaces WHERE name = ?`, iface.Name)
	row.Scan(&ifaceID)

	for _, addr := range iface.Addresses {
		_, err = s.db.Exec(
			`INSERT OR IGNORE INTO server_addresses (address, mask) VALUES (?, ?)`,
			addr.Address,
			addr.Mask,
		)
		if err != nil {
			return err
		}

		var addrID int
		row := s.db.QueryRow(`SELECT id FROM server_addresses WHERE address = ? AND mask = ?`, addr.Address, addr.Mask)
		row.Scan(&addrID)

		_, err = s.db.Exec(
			`INSERT OR IGNORE INTO server_interface_addresses (interface_id, address_id) VALUES (?, ?)`,
			ifaceID,
			addrID,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ServiceDB) getIfaceFromID(id int) (*DBIface, error) {
	row := s.db.QueryRow(
		`SELECT name, create_on_startup FROM server_interfaces WHERE id = ?`,
		id,
	)

	iface := DBIface{}

	err := row.Scan(&iface.Name, &iface.CreateOnStartup)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(
		`SELECT address, mask
			FROM       server_addresses           sa
			INNER JOIN server_interface_addresses sia ON sia.address_id   = sa.id
			INNER JOIN server_interfaces          si  ON sia.interface_id = si.id
			WHERE si.name = ?`,
		iface.Name,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	addrs := []wireconnect.Address{}
	for rows.Next() {
		var address net.IP
		var mask net.IPMask

		if err := rows.Scan(&address, &mask); err != nil {
			return nil, err
		}

		addrs = append(addrs, wireconnect.Address{Address: address, Mask: mask})
	}

	iface.Addresses = addrs

	return &iface, nil
}
