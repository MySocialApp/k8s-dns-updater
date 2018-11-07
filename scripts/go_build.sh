#!/usr/bin/env bash

GO_VERSION=1.11.2

if [ -d /vagrant_data ] ; then
    echo "[+] Installing go $GO_VERSION"
    cd /tmp
    wget -q https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz
    tar -xf go${GO_VERSION}.linux-amd64.tar.gz
    mv go /usr/local

    export GOROOT=/usr/local/go
    export GOPATH=$HOME
    export PATH=$GOPATH/bin:$GOROOT/bin:$PATH

    echo "[+] Copy project to $GOPATH"
    mkdir -p $GOPATH/src/github.com/MySocialApp
    cp -Rf /vagrant_data $GOPATH/src/github.com/MySocialApp/k8s-dns-updater

    cd /vagrant_data
fi

go build