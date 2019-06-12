# Install developer tools
sudo yum groupinstall -y "Development Tools"

# Setup aliases
echo "alias l='ls -alt'" >> ~/.bashrc

# Install go
aws s3 cp s3://mailctl-setup/install-go.sh .
sudo bash install-go.sh
bash ~/.bashrc

# Install go packages
go get github.com/danielhavir/mailctl
