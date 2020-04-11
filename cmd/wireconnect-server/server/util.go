package server

import (
	"bufio"
	"fmt"
	"math/bits"
	"net"
	"os"
	"syscall"

	"github.com/sector-f/wireconnect"
	"github.com/sector-f/wireconnect/cmd/wireconnect-server/database"
	"golang.org/x/crypto/ssh/terminal"
)

func cidr(mask net.IPMask) uint {
	var cidrmask uint

	for _, b := range mask {
		cidrmask += uint(bits.OnesCount(uint(b)))
	}

	return cidrmask
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

	return s.db.AddUser(database.User{Username: username, Password: password, IsAdmin: true})
}

func (s *Server) makeFirstIface() error {
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

		addresses, err = database.CidrList(addr)
		if err != nil {
			continue
		}
		break
	}

	return s.db.AddIface(
		database.DBIface{
			Name:      "wireconnect0",
			Addresses: addresses,
		},
	)
}
