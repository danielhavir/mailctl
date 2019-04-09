# Install developer tools
sudo yum groupinstall -y "Development Tools"

# Setup aliases
echo "alias l='ls -alt'" >> ~/.bashrc

# Install go
aws s3 cp s3://mailctl-setup/install-go.sh .
sudo bash install-go.sh
source ~/.bashrc

# Install go packages
mkdir $GOPATH/bin
go get -u github.com/flashmob/go-guerrilla
go get github.com/flashmob/maildir-processor
curl https://glide.sh/get | sh

# Prepare
cd ~/go/src/github.com/flashmob/go-guerrilla
glide install
cp goguerrilla.conf.sample goguerrilla.conf.json

