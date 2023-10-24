# GoDaddy DDNS Server & Client
Inspired By [https://github.com/proprietary/godaddns](https://github.com/proprietary/godaddns)


This repo contains code for the server and client. 

The server receives the dns change requests from the clients together with credentials. Credentials are from a Cosmos-based chain secp256k1 key set.

The client periodically monitors the dns record and determines if a change request is required.

## Envs
### Client
```
SERVER_URL=
NODE_ID=
PASSWORD=
IP_PROVIDER=
PRIV_KEY_PATH=
```
### Server
```
GODADDY_KEY=
GODADDY_SECRET=
DOMAIN=
```

## Docker Compose Use
Refer to files in `deploy`.