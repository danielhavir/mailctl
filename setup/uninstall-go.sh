#!/bin/bash
# Resource: https://golang.org/doc/install#uninstall

# The following needs sudo priviliges:
rm -rf /usr/local/go

tput setaf 3; tput bold
echo """You should also remove the Go bin directory from your PATH environment variable."""
echo """Under Linux and FreeBSD you should edit /etc/profile or $HOME/.profile."""
echo """If you installed Go with the macOS package then you should remove the /etc/paths.d/go file."""
tput sgr0
