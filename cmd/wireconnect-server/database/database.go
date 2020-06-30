package database

import (
	"database/sql"
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
	is_admin BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS peers (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	address INTEGER NOT NULL,
	mask INTEGER NOT NULL,
	endpoint_address INTEGER NOT NULL,
	server_interface_id INTEGER NOT NULL,
	user_id INTEGER NOT NULL,
	FOREIGN KEY(server_interface_id) REFERENCES server_interfaces(id),
	FOREIGN KEY(user_id) REFERENCES users(id),
	UNIQUE(name, user_id)
);`,
	)

	return err
}
