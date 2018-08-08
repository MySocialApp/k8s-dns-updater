#!/usr/bin/env bash

echo "Starting Kubernetes DNS Updater"

while IFS='=' read -r name value ; do
  if [[ $name == 'KDU_'* ]]; then
    val="${!name}"
    echo "$name $val"
    field=$(echo $name | sed -r 's/KDU_(.+)/\1/')
    sed -ri "s/($field):.*/\1: $val/" "/etc/k8s-dns-updater/config.yaml"
  fi
done < <(env)

/usr/bin/k8s-dns-updater