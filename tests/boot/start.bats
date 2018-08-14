#!/usr/bin/env bats

load ../k8s-euft/env
load common

@test "Label nodes" {
    label_nodes
}

@test "Ensure nodes has correct labels" {
    num_nodes_are_labeled_as_node
}

@test "Ensure k8s-dns-updater to populate DNS" {

}
