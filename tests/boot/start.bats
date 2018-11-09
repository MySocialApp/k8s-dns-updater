#!/usr/bin/env bats

load ../k8s-euft/env
load common

@test "Label nodes" {
    label_nodes
}

@test "Create DNS nodes records" {
    #create_dns_hosts_records
}

@test "Ensure there is no other running k8s-dns-updater" {
    kill_app
}

@test "Start k8s-dns-updater" {
    run_app
    status_app
}
