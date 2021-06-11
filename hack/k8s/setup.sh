#!/bin/bash
set -euo pipefail

ready_cluster() {
    # The cluster takes a few seconds to respond after creation.
    while kubectl get pods -A --output=json |
        jq -e '.items | length == 0' >/dev/null; do
        sleep 1
    done

    kubectl wait --namespace kube-system \
        --for=condition=ready pod \
        --selector=tier=control-plane \
        --timeout=90s
    kubectl -n kube-system get pods
}

cluster() {
    kind create cluster
    ready_cluster
}

add_repos() {
    helm repo add \
        prometheus-community \
        https://prometheus-community.github.io/helm-charts

    helm repo add kube-state-metrics https://kubernetes.github.io/kube-state-metrics

    helm repo add \
        grafana \
        https://grafana.github.io/helm-charts

    helm repo update
}

cluster
add_repos

helm upgrade --install \
    prometheus \
    prometheus-community/prometheus \
    --values ./prometheus-values.yaml

helm upgrade --install \
    grafana \
    grafana/grafana \
    --values ./grafana-values.yaml

# kubectl port-forward service/prometheus-server 9090:80
# kubectl port-forward service/grafana 3000:80
