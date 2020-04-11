package server

import (
	"bufio"
	"database/sql"
	"fmt"
	"net"
	"os"
	"syscall"

	"github.com/sector-f/wireconnect"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh/terminal"
)

type ServiceDB struct {
	db *sql.DB
}

func New(db *sql.DB) (*ServiceDB, error) {
	s := ServiceDB{db}

	err := s.initDB()
	if err != nil {
		return &s, err
	}

	return &s, nil
}

type User struct {
	Username    string
	Password    []byte
	PeerConfigs []PeerConfig
	IsAdmin     bool
}

type DBIface struct {
	Name            string
	CreateOnStartup bool
	Addresses       []wireconnect.Address
}

type PeerConfig struct {
	Name    string
	Address wireconnect.Address
	DBIface *DBIface
}

func (s *ServiceDB) initDB() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS server_interfaces (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT UNIQUE NOT NULL,
	create_on_startup BOOLEAN NOT NULL DEFAULT true,
	UNIQUE(name)
);

CREATE TABLE IF NOT EXISTS server_addresses (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	address INTEGER NOT NULL,
	mask INTEGER NOT NULL,
	UNIQUE(address, mask)
);

CREATE TABLE IF NOT EXISTS server_interface_addresses (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	interface_id INTEGER NOT NULL,
	address_id INTEGER NOT NULL,
	FOREIGN KEY(interface_id) REFERENCES server_interface_addresses(id),
	FOREIGN KEY (address_id)  REFERENCES addresses(id),
	UNIQUE(interface_id, address_id)
);

CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT UNIQUE NOT NULL,
	password TEXT NOT NULL,
	is_admin BOOLEAN NOT NULL DEFAULT false,
);

CREATE TABLE IF NOT EXISTS peers (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	address INTEGER NOT NULL,
	mask INTEGER NOT NULL,
	server_interface_id INTEGER NOT NULL,
	user_id INTEGER NOT NULL,
	FOREIGN KEY(server_interface_id) REFERENCES server_interfaces(id),
	FOREIGN KEY(user_id) REFERENCES users(id),
	UNIQUE(name, user_id)
);`,
	)

	return err
}

func (s *ServiceDB) Authenticate(username, password string) error {
	var dbPass string

	row := s.db.QueryRow(`SELECT password FROM users WHERE username = ?`, username)
	switch err := row.Scan(&dbPass); err {
	case sql.ErrNoRows:
		return err
	case nil:
		return bcrypt.CompareHashAndPassword([]byte(dbPass), []byte(password))
	default:
		return err
	}
}

func (s *ServiceDB) IsAdmin(username string) (bool, error) {
	var isAdmin bool

	row := s.db.QueryRow(`SELECT is_admin FROM users WHERE username = ?`, username)
	switch err := row.Scan(&isAdmin); err {
	case sql.ErrNoRows:
		return false, err
	case nil:
		return isAdmin, nil
	default:
		return false, err
	}
}

func (s *ServiceDB) AddUser(user User) error {
	hashedPw, err := bcrypt.GenerateFromPassword(user.Password, bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		`INSERT INTO users (username, password, is_admin) VALUES (?, ?, ?)`,
		user.Username,
		string(hashedPw),
		user.IsAdmin,
	)

	return err
}

func (s *ServiceDB) UserCount() (uint, error) {
	var count uint

	row := s.db.QueryRow(`SELECT COUNT(*) FROM users`)
	switch err := row.Scan(&count); err {
	case sql.ErrNoRows, nil:
		return count, nil
	default:
		return 0, err
	}
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

// TODO: Move this outside this package
func (s *ServiceDB) makeFirstUser() error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Creating initial admin user")

	fmt.Print("Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	username = username[:len(username)-1]

	fmt.Print("Password: ")
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return err
	}
	fmt.Println()

	return s.AddUser(User{Username: username, Password: password, IsAdmin: true})
}

// TODO: Move this outside this package
func (s *ServiceDB) makeFirstIface() error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Creating initial wireguard interface.")

	var addresses []wireconnect.Address

	for {
		fmt.Println("Please enter a comma-seperated list of IP addresses in CIDR notation.")

		fmt.Print("> ")
		addr, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		addr = addr[:len(addr)-1]

		addresses, err = CidrList(addr)
		if err != nil {
			continue
		}
		break
	}

	return s.AddIface(
		DBIface{
			Name:      "wireconnect0",
			Addresses: addresses,
		},
	)
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

func (s *ServiceDB) Ifaces() ([]DBIface, error) {
	ifaces := []DBIface{}

	rows, err := s.db.Query(`SELECT name, create_on_startup FROM server_interfaces`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var cos bool
		if err := rows.Scan(&name, &cos); err != nil {
			return nil, err
		}
		ifaces = append(ifaces, DBIface{Name: name, CreateOnStartup: cos})
	}

	for i, iface := range ifaces {
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

		ifaces[i].Addresses = addrs
	}

	return ifaces, nil
}
