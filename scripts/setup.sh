#!/bin/bash
set -e
echo "Luxior OSINT Setup"

if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    sudo apt-get update
    sudo apt-get install -y build-essential golang rustc cargo nodejs npm python3 python3-pip redis-server postgresql
elif [[ "$OSTYPE" == "darwin"* ]]; then
    brew install go rust node python redis postgresql
fi

pip3 install -r requirements.txt
npm install

echo "Setup complete"

