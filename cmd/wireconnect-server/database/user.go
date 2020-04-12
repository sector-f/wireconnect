package database

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username    string
	Password    []byte
	PeerConfigs []PeerConfig
	IsAdmin     bool
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
