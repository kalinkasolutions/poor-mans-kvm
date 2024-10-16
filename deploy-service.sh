#!/bin/sh


GO_CMD=$(which go)

if [ -z "$GO_CMD" ]; then
    echo "Go is not installed or not in PATH"
    exit 1
fi

$GO_CMD build

sudo systemctl stop poor-mans-kvm.service

cp config.json /home/kalinka/.config/poormanskvm/config.json
sudo cp poor-mans-kvm /usr/local/bin/poor-mans-kvm
sudo cp poor-mans-kvm.service /etc/systemd/system/poor-mans-kvm.service

sudo systemctl daemon-reload
sudo systemctl enable poor-mans-kvm.service
sudo systemctl start poor-mans-kvm.service