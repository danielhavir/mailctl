# mailctl client

Server listener for sending / receiving encrypted messages

## Build
* `go build *.go -o ./mailctl`

## Run server:
* `./mailctl -storage-path=<storage-path | OPTIONAL> -address=<address or port | OPTIONAL>`

### Examples:

* `./mailctl -storage-path=$HOME/.mailctl`
* `./mailctl -address=":1991"`
