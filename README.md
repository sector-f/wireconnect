Flow
====
1. HTTP server listens on port
2. Client connects to server, sends POST (as JSON?) to /connect using HTTP Basic Auth
4. Server validates username/password
	* Using SQL database (probably sqlite)
	* Gets user's IP address from database
5. (Assuming successful authentication) server sends reply with following information (as JSON?):
	* Server's public wireguard key (randomly generated at server startup)
	* IP address that client must use for its wireguard interface
	* Server's IP address that shall be used as client's peer's address
6. Server adds peer using client's public key (using the IP address that client connected from as peer address)
7. Client creates interface using server's IP address and public key

HTTP Routes
===========
* /connect
* /disconnect
