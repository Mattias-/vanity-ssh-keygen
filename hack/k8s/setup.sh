#!/bin/bash
set -euo pipefail

helm repo add \
    prometheus-community \
    https://prometheus-community.github.io/helm-charts

helm repo add \
    grafana \
    https://grafana.github.io/helm-charts

helm repo update

helm install \
    prometheus \
    prometheus-community/prometheus \
    --values ./prometheus-values.yaml

helm install \
    grafana \
    grafana/grafana \
    --values ./grafana-values.yaml

# kubectl port-forward service/prometheus-server 9090:80
# kubectl port-forward service/grafana 3000:80
