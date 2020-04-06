package server

import (
	"bufio"
	"database/sql"
	"fmt"
	"net"
	"os"
	"syscall"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh/terminal"
)

type User struct {
	Username string
	Password []byte
	IsAdmin  bool
}

type DBIface struct {
	Name    string
	Address string
	Mask    []byte
}

func (s *Server) initDB() error {
	_, err := s.db.Exec(
		`CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT UNIQUE NOT NULL,
	password TEXT NOT NULL,
	is_admin BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS interfaces (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	address TEXT NOT NULL,
	mask INTEGER NOT NULL
);`,
	)

	return err
}

func (s *Server) authenticate(username, password string) error {
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

func (s *Server) isAdmin(username string) (bool, error) {
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

func (s *Server) addUser(user User) error {
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

func (s *Server) userCount() (uint, error) {
	var count uint

	row := s.db.QueryRow(`SELECT COUNT(*) FROM users`)
	switch err := row.Scan(&count); err {
	case sql.ErrNoRows, nil:
		return count, nil
	default:
		return 0, err
	}
}

func (s *Server) ifaceCount() (uint, error) {
	var count uint

	row := s.db.QueryRow(`SELECT COUNT(*) FROM interfaces`)
	switch err := row.Scan(&count); err {
	case sql.ErrNoRows, nil:
		return count, nil
	default:
		return 0, err
	}
}

func (s *Server) makeFirstUser() error {
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

	return s.addUser(User{Username: username, Password: password, IsAdmin: true})
}

func (s *Server) makeFirstIface() error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Creating initial wireguard interface")

	fmt.Print("IP Address: ")
	addr, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	addr = addr[:len(addr)-1]

	ip, network, err := net.ParseCIDR(addr)
	if err != nil {
		return err
	}

	return s.addIface(DBIface{Name: "wireconnect0", Address: ip.String(), Mask: network.Mask})
}

func (s *Server) addIface(iface DBIface) error {
	_, err := s.db.Exec(
		`INSERT INTO interfaces (name, address, mask) VALUES (?, ?, ?)`,
		iface.Name,
		iface.Address,
		iface.Mask,
	)

	return err
}
