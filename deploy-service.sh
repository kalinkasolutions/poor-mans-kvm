#!/bin/sh

GO_CMD=$(command -v go)

if [ -z "$GO_CMD" ]; then
    echo "Go is not installed or not in PATH"
    exit 1
fi

$GO_CMD build || exit 1

sudo systemctl stop poor-mans-kvm.service 2>/dev/null || true

sudo mkdir -p /etc/poormanskvm
if [ -f /etc/poormanskvm/config.json ]; then
    printf "Config already exists at /etc/poormanskvm/config.json. Overwrite? [y/N] "
    read -r answer
    case "$answer" in
        [yY]) sudo cp config.json /etc/poormanskvm/config.json ;;
        *) echo "Keeping existing config." ;;
    esac
else
    sudo cp config.json /etc/poormanskvm/config.json
fi

sudo cp poor-mans-kvm /usr/local/bin/poor-mans-kvm
sudo cp poor-mans-kvm.service /etc/systemd/system/poor-mans-kvm.service

sudo systemctl daemon-reload
sudo systemctl enable poor-mans-kvm.service
sudo systemctl start poor-mans-kvm.service