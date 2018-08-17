#!/usr/bin/env bash

set -e
echo "Updating all version files from current app version..."

go build
APP_VERSION=$(./k8s-dns-updater version)
APP_VERSION_WITHOUT_V=$(echo ${APP_VERSION} | sed 's/v//')

for filename in README.md kubernetes/values.yaml ; do
    sed -ri "s/(KduImageVersion: ).*/\1${APP_VERSION}/g" ${filename}
done

sed -ri "s/^(ARG version=).*/\1${APP_VERSION_WITHOUT_V}/g" Dockerfile
sed -ri "s/(appVersion: ).*/\1\"${APP_VERSION}\"/g" kubernetes/Chart.yaml

echo "Done"