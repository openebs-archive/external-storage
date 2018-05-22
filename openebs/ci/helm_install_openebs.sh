#!/usr/bin/env bash

kubectl -n kube-system create sa tiller
kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller
kubectl -n kube-system patch deploy/tiller-deploy -p '{"spec": {"template": {"spec": {"serviceAccountName": "tiller"}}}}'

#Replace this with logic to wait till helm is initialized
sleep 30
kubectl get pods --all-namespaces

helm repo add openebs-charts https://openebs.github.io/charts/
helm repo update
helm install openebs-charts/openebs --namespace default --set apiserver.imageTag="ci",apiserver.replicas="1",provisioner.imageTag="ci",provisioner.replicas="1",jiva.replicas="1"

#Replace this with logic to wait/verify openebs control plane is initialized
sleep 30
kubectl get pods --all-namespaces
kubectl get svc --all-namespaces
