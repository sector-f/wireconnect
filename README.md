Flow
====
1. HTTP server listens on port
2. Server creates wireguard interface with no peers
3. Client connects to server, sends POST (with JSON body) to /connect using HTTP Basic Auth
4. Server validates username/password
	* Using SQL database (probably sqlite)
	* Gets user's wireguard IP address from database
5. (Assuming successful authentication) server sends reply with following information (as JSON):
	* Server's public wireguard key (randomly generated at server startup)
	* IP address that client must use for its wireguard interface
	* Server's IP address that shall be used as client's peer's address
6. Server adds peer to wireguard interface using client's public key (with the IP address that client connected from being used as the peer address)
7. Client creates wireguard interface using server's IP address and public key

To-Do
=====
* Server:
	* Force TLS
	* Add routes for adding/modifying/deleting users
	* Add code to configure wireguard interface(s)
	* Make responses use JSON
	* Add block time to rate-limit ban list (if feasible)

* Client:
	* Implement everything

HTTP Routes
===========
* /connect
* /disconnect
