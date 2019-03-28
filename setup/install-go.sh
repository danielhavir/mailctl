#!/bin/bash
# Resource 1: https://golang.org/doc/install#install
# Resource 2: https://alexatnet.com/install-go-on-raspberry-pi/

version="1.12.1"
machine=`uname -m`

bold=$(tput bold)
yellow=$(tput setaf 3)
green=$(tput setaf 2)
normal=$(tput sgr0)

if [ $machine == armv7l ]
then
    echo "Arch: $machine"
    tarfile="go$version.linux-armv6l.tar.gz"
elif [ $machine == x86_64 ]
then
    os=`uname`
    if [ $os == "Linux" ]
    then
        echo "Arch: $os $machine"
        tarfile="go$version.linux-amd64.tar.gz"
    elif [ $os == "Darwin" ]
    then
        echo "Arch: $os $machine"
        tarfile="go$version.darwin-amd64.tar.gz"
    fi
else
    tput setaf 1; tput bold; echo "Unknown CPU architecture $machine"; tput sgr0
    echo "Visit https://golang.org/dl/"
    exit 0
fi
echo "####"

url="https://dl.google.com/go/$tarfile"
echo "${bold}Downloading${normal} $url"
wget $url
echo "####"

# The following command needs sudo privileges
echo "Extracting $tarfile, sudo privileges required:"
sudo tar -C /usr/local -xzf $tarfile
echo "####"

echo "${green}Go successfully installed${normal}"
echo "Appending ${yellow}${bold}export PATH=\$PATH:/usr/local/go/bin${normal} to ${yellow}${bold}~/.bashrc${normal}"

echo "# Go" >> ~/.bashrc
echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc

