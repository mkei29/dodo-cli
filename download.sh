#!/bin/bash

set -e

VERSION="0.0.7"
OS=$(uname -s)
LOCAL_ARCH=$(uname -m)

case $LOCAL_ARCH in
  x86_64|amd64)
    ARCH=x86_64
    ;;
  armv8|aarch64|arm64)
    ARCH=arm64
    ;;
  i386)
    ARCH=i386
    ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# Download the binary
URL="https://github.com/toritoritori29/dodo-cli/releases/download/$VERSION/dodo-cli_${OS}_${ARCH}.tar.gz"
echo "[1/2] Downloading dodo-cli from $URL ..."
curl -sLO $URL
filename="dodo-cli_${OS}_${ARCH}.tar.gz"

echo "[2/2] Extracting dodo-cli from the archive ..."
tar -xzf $filename
rm $filename
chmod +x dodo-cli

printf "\ndodo-cli %s Download Complete!" $VERSION
printf "\ndodo-cli has been successfully downloaded into the current directory."
printf "\n"
printf "\nTo install dodo-cli on your system, move the executable to a directory in your PATH."
printf "\n"
printf "\n\tmv ./dodo-cli /usr/local/bin"
printf "\n"
printf "\nAfter installation, verify the installation by running 'dodo-cli --version'."
printf "\nNeed more information? Please visit https://www.dodo-doc.com."
exit 0