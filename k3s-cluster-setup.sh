#!/bin/bash
# K3s cluster initialization script
# Usage: bash k3s-cluster-setup.sh

set -e

# Install K3s (single node, for demo/dev)
curl -sfL https://get.k3s.io | sh -

# Wait for K3s to be ready
sleep 10

# Show node status
sudo k3s kubectl get nodes

# Install Helm (optional, for charts)
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Print kubeconfig location
echo "Kubeconfig: /etc/rancher/k3s/k3s.yaml"

# Reminder for multi-node setup
echo "For multi-node, see: https://rancher.com/docs/k3s/latest/en/networking/"
