# mailctl client

CLI for sending / receiving encrypted messages

## Build
* `go build *.go -o ./mailctl`

## Example Usage

### Configure:
* `./mailctl configure -config-path=<config-path | OPTIONAL>`

### Send message:
* `./mailctl send -rcpt=<user>@<org> -file=<file> -subject=<subject | OPTIONAL>`

### List unread messages:
* `./mailctl list -config-path=<config-path | OPTIONAL>`

### Receive message:
* `./mailctl recv -message-id=<message_id> -config-path=<config-path | OPTIONAL>`

### Help:
* `./mailctl help`
