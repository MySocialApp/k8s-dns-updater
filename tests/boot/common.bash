#!/usr/bin/env bash

#set -x

label_nodes() {
    for i in $(kubectl get no -l node-role.kubernetes.io/master!= | awk '/kube-node-/{ print $1 }') ; do
        kubectl label node $i node-role.kubernetes.io/node=true --overwrite
    done
}

load_clouflare_config() {
    # Load from config file or variable env
    config_file="../../config.yaml"
    if [ -f $config_file ] ; then
        export cf_zoneid=$(awk -F'"' '/ZoneId/{ print $2 }' $config_file)
        export cf_zonename=$(awk -F'"' '/ZoneName/{ print $2 }' $config_file)
        export cf_name=$(awk '/ Name:/{ print $2 }' $config_file)
        export cf_email=$(awk -F'"' '/Email/{ print $2 }' $config_file)
        export cf_key=$(awk -F'"' '/Key/{ print $2 }' $config_file)
    fi

    # Check if vars are defined
    for i in $cf_zoneid $cf_name $cf_email $cf_key ; do
        if [ $i == "" ] ; then
            echo "Error, variable $i is not set"
            exit 1
        fi
    done
}

get_cloudflare_record_id() {
    record=$1
    return_id_or_content=$2
    to_return=${return_id_or_content:-content}
    load_clouflare_config

    curl -s "https://api.cloudflare.com/client/v4/zones/${cf_zoneid}/dns_records?name=${record}&page=1&per_page=50&type=A" \
      -H "X-Auth-Email: ${cf_email}" \
      -H "X-Auth-Key: ${cf_key}" \
      -H "Content-Type: application/json" | jq --raw-output ".result[].${to_return}"
}

get_cloudflare_registered_nodes() {
    return_id_or_content=$1
    to_return=${return_id_or_content:-content}
    load_clouflare_config

    if [ "$to_return" == 'id' ] ; then
        nodes=$(get_cloudflare_record_id ${cf_name} 'id')
    else
        nodes=$(get_cloudflare_record_id ${cf_name} 'content')
    fi

    if [ $(echo $nodes | wc -l) == 0 ] ; then
        echo "Error while trying to get records from DNS name ${cf_name} on Cloudlfare"
        exit 1
    fi

    echo $nodes
}

clear_cloudflare_entries() {
    load_clouflare_config

    # Delete round robin entries
    rr_list=$(get_cloudflare_registered_nodes 'id')
    for id in `echo $rr_list` ; do
        curl -s -X DELETE "https://api.cloudflare.com/client/v4/zones/${cf_zoneid}/dns_records/${id}" \
          -H "X-Auth-Email: ${cf_email}" \
          -H "X-Auth-Key: ${cf_key}" \
          -H "Content-Type: application/json"
    done

    # Delete nodes
    for node_name in $(kubectl get no -l node-role.kubernetes.io/master!= | awk '/kube-node-/{ print $1 }') ; do
        node_id=$(get_cloudflare_record_id ${node_name}.${cf_zonename} 'id')
        curl -s -X DELETE "https://api.cloudflare.com/client/v4/zones/${cf_zoneid}/dns_records/${node_id}" \
          -H "X-Auth-Email: ${cf_email}" \
          -H "X-Auth-Key: ${cf_key}" \
          -H "Content-Type: application/json"
    done
}

#get_cloudflare_record_id bla.4tech.io
#get_cloudflare_registered_nodes 'content'
#get_cloudflare_registered_nodes id
#clear_cloudflare_entries