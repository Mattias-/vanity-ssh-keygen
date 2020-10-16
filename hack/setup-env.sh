#!/bin/bash
set -euo pipefail

apt-get update -y
apt-get install -y git curl htop snapd
snap install go --classic

git clone https://github.com/Mattias-/vanity-ssh-keygen.git
