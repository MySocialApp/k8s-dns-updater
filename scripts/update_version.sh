#!/usr/bin/env bash

set -e

echo "Updating all version files from current app version..."

go build
VERSION=$(./k8s-dns-updater version)
VERSION_WITHOUT_V=$(echo ${VERSION} | sed 's/v//')

for filename in README.md kubernetes/values.yaml ; do
    sed -ri "s/(KduImageVersion: ).+/\1${VERSION}/g" ${filename}
done

sed -ri "s/(ARG version=).+/\1${VERSION_WITHOUT_V}/g" Dockerfile
sed -ri "s/(appVersion: ).+/\1\"${VERSION}\"/g" kubernetes/Chart.yaml

echo "Done"