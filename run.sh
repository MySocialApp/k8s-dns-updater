#!/usr/bin/env bash

echo "Starting Kubernetes DNS Updater"

while IFS='=' read -r name value ; do
  if [[ $name == 'KDU_'* ]]; then
    val="${!name}"
    if [ "$name" != "KDU_Key" ] && [ "$name" != "KDU_Email" ] ; then
      echo "$name $val"
    fi
    field=$(echo $name | sed -r 's/KDU_(.+)/\1/')
    sed -ri "s/($field):.*/\1: $val/" "/etc/k8s-dns-updater/config.yaml"
  fi
done < <(env)

/usr/bin/k8s-dns-updater