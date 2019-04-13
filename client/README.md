# mailctl client

CLI for sending / receiving encrypted messages

## Build
* `go build *.go -o mailctl`

## Example Usage

### Configure:
* `./mailctl configure --config-file=<config-file | OPTIONAL>`

### Send message:
* `./mailctl send --rcpt=<user>@<org> --file=<file> -subject=<subject | OPTIONAL>`

### List unread messages:
* `./mailctl list --config-file=<config-file | OPTIONAL>`

### Receive message:
* `./mailctl recv --message-id=<message_id> --config-file=<config-file | OPTIONAL>`

### Help:
* `./mailctl help`
