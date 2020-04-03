package main

import (
	"github.com/sector-f/wireconnect"
)

type mockAuth struct{}

func (m mockAuth) Auth(req wireconnect.Request) (wireconnect.Reply, error) {
	reply := wireconnect.Reply{
		Error:         "",
		PublicKey:     "abc123",
		ClientAddress: "10.0.0.2",
		ServerAddress: "192.168.0.1",
	}

	return reply, nil
}
