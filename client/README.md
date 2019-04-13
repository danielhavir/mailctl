# mailctl client

## Build
* `go build *.go -o client.out`

## Run
Subcommands:

* **configure** (set up configuration): `./client.out configure --config-file=<path to config file | OPTIONAL>`
* **send**: `./client.out send --rcpt=< 'user'@'org' | REQUIRED> --file=<path to file to be send | REQUIRED> --subject=<subject line | OPTIONAL> --config-file=<path to config file | OPTIONAL>`
* **recv**: `./client.out recv --message-id=<message id | REQUIRED> --config-file=<path to config file | OPTIONAL>`
* **list** (list unread messages): `./client.out list --config-file=<path to config file | OPTIONAL>`

### Help
Run `./client.out -h`
