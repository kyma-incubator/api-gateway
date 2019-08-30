# Api-Gateway Controller (name to be changed)

## Overview

The API Gateway Controller manages Istio authentication Policies, VirtualServices and Oathkeeper Rule. The controller allows to expose services using Gate resources. It operates on `gate.gateway.kyma-project.io` CustomResourceDefinition (CRD) resources.

## Prerequisites

- recent version of Go language with support for modules (e.g: 1.12.6)
- make
- kubectl
- kustomize
- access to K8s environment: minikube or a remote K8s cluster

## Details

### Run the controller locally

- `start minikube`
- `make build` to build the binary and run tests
- `eval $(minikube docker-env)`
- `make build-image` to build a docker image
- export `OATHKEEPER_SVC_ADDRESS`, `OATHKEEPER_SVC_PORT` and `JWKS_URI` variables
- `make deploy` to deploy controller to the minikube

### Use command-line flags

| Name | Required | Description | Possible values |
|------|----------|-------------|-----------------|
| **oathkeeper-svc-address** | Yes | Used to set ory oathkeeper-proxy service address. | ` ory-oathkeeper-proxy.kyma-system.svc.cluster.local` |
| **oathkeeper-svc-port** | Yes | Used to set ory oathkeeper-proxy service port. | `4455` |
| **jwks-uri** | Yes | Used to set default jwksUri in the Policy. | any string |

### Example CR structure

Valid examples of Gate CR can be found in `config/samples` catalog. 