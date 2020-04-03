package wireconnect

type AuthProvider interface {
	Auth(req Request) (Reply, error)
}

type Request struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	PublicKey string `json:"public_key"`
}

type Reply struct {
	Error         string `json:"error"`
	PublicKey     string `json:"public_key"`
	ClientAddress string `json:"client_address"`
	ServerAddress string `json:"server_address"`
}

type User struct {
	Username     string
	Password     string
	Address      string
	PresharedKey string
}
