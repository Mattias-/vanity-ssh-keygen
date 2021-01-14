#!/bin/bash
set -euo pipefail

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
