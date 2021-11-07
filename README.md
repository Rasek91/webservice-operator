# Webservice-operator

Webservice-operator is an operator for deploying a webservice on Kubernetes with TLS encrypted external NGINX ingress controller.

## Features

Webservice-operator is doing the following steps:

- Create a certificate with issuer `.spec.issuer` and for hostname `.spec.host`
- Create a deployment with replicas `.spec.replicas`, with resources `.spec.resources` and with a container from image `.spec.image` and listen on port `.spec.containerPort` the default port is 80
- Create a service where the deployment is available
- Create an ingress controller which is using the service as backend, using the certificate and listening on hostname `.spec.host`

## Install

### Prerequisites

Change the `.spec.acme.email` in the samples configuration files: [letsencrypt_production.yaml](config/samples/letsencrypt_production.yaml) and [letsencrypt_staging.yaml](config/samples/letsencrypt_staging.yaml), before applying them to your cluster.

1. Install NGINX Ingress Controller

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.0.4/deploy/static/provider/cloud/deploy.yaml
#use this deployment if you have kind cluster
#more info: https://kind.sigs.k8s.io/docs/user/ingress/
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
```

2. Install cert-manager

```bash
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.6.0/cert-manager.yaml
```

3. Add Issuer to the cluster

```bash
#Let's Encrypt stating environment, good for testing
kubectl apply -f config/samples/letsencrypt_staging.yaml
#Let's Encrypt production environment, the generated cert will be accepted by the browsers
kubectl apply -f config/samples/letsencrypt_production.yaml
```

### Operator

1. Set `KUBECONFIG` pointing towards your cluster

2. Deploy the operator in the `webservice-operator-system` namespace to your cluster

```bash
make deploy
```

3. Change the sample configuration file: [webservice_v1_webapp.yaml](config/samples/webservice_v1_webapp.yaml) and applying it to your cluster

```bash
kubectl apply -f config/samples/webservice_v1_webapp.yaml
```
