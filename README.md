# Api-Gateway Controller (name to be changed)

## Overview

The API Gateway Controller manages Istio VirtualServices and Oathkeeper Rule. The controller allows to expose services using instances of the `apirule.gateway.kyma-project.io` custom resource (CR).

## Prerequisites

- recent version of Go language with support for modules (e.g: 1.12.6)
- make
- kubectl
- kustomize
- access to K8s environment: minikube or a remote K8s cluster

## Details

### Deploy to the cluster
Deploys the oficially released Controller version to the cluster

- ensure the access to a Kubernetes cluster is configured in `~/.kube/config`
- `make install` to install necessary Custom Resource Definitions
- export `OATHKEEPER_SVC_ADDRESS`, `OATHKEEPER_SVC_PORT`, `JWKS_URI` and `DOMAIN_WHITELIST` variables
- `make deploy` to deploy controller

### Run the controller locally
This procedure is the fastest way to run the Controller, useful for development purposes

- start Minikube or ensure the access to a Kubernetes cluster is configured in `~/.kube/config`
- `make install` to install necessary Custom Resource Definitions
- export `OATHKEEPER_SVC_ADDRESS`, `OATHKEEPER_SVC_PORT`, `JWKS_URI` and `DOMAIN_WHITELIST` variables
- `go run main.go --jwks-uri="$JWKS_URI" --oathkeeper-svc-address="$OATHKEEPER_SVC_ADDRESS" --oathkeeper-svc-port=$OATHKEEPER_SVC_PORT --domain-whitelist=$DOMAIN_WHITELIST`

### Deploy a custom Controller build to the local Minikube cluster
This procedure is useful to test your own Controller build end-to-end in a local Minikube cluster.

- start Minikube
- `make build` to build the binary and run tests
- `eval $(minikube docker-env)`
- `make build-image` to put the docker image inside running Minikube
- `make install` to install necessary Custom Resource Definitions
- export `OATHKEEPER_SVC_ADDRESS`, `OATHKEEPER_SVC_PORT`, `JWKS_URI` and `DOMAIN_WHITELIST` variables
- `make deploy-dev` to deploy controller

### Use command-line flags

| Name | Required | Description | Example values |
|------|----------|-------------|-----------------|
| **oathkeeper-svc-address** | yes | ory oathkeeper-proxy service address. | `ory-oathkeeper-proxy.kyma-system.svc.cluster.local` |
| **oathkeeper-svc-port** | yes | ory oathkeeper-proxy service port. | `4455` |
| **jwks-uri** | yes | default jwksUri in the Policy. | any string |
| **service-blacklist** | no | list of services to be blacklisted | `kubernetes.default` <br> `kube-dns.kube-system` |
| **domain-whitelist** | yes | list of domains that can be exposed | `kyma.local` <br> `foo.bar` |
| **cors-allow-origin**  | no | comma-separated list of allowed origins | `*`, `https://developer.org` |
| **cors-allow-methods** | no | comma-separated list of allowed methods | `GET,POST,DELETE` |
| **cors-allow-headers** | no | comma-separated list of allowed headers | `Authorization,Content-Type` |

## Custom Resource

The `apirule.gateway.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format the API Gateway Controller listens for. To get the up-to-date CRD and show
the output in the `yaml` format, run this command:
```
kubectl get crd apirule.gateway.kyma-project.io -o yaml
```

### Sample custom resource

This is a sample custom resource (CR) that the API-gateway listens for to expose a service.

```
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: jwt-all-with-scopes
spec:
  gateway: kyma-gateway.kyma-system.svc.cluster.local
  service:
    name: foo-service
    port: 8080
    host: foo.bar
  rules:
    - path: /.*
      methods: ["GET"]
      mutators: []
      accessStrategy:
        - handler: jwt
          config:
            trusted_issuers: ["http://dex.kyma.local"]
            required_scope: ["foo", "bar"]

```

This table lists all the possible parameters of a given resource together with their descriptions:

| Field   |      Mandatory      |  Description |
|:---|:---:|:---|
| **metadata.name** |    **YES**   | Specifies the name of the exposed API |
| **spec.gateway** | **YES** | Specifies Istio Gateway. |
| **spec.service.name**, **spec.service.port** | **YES** | Specifies the name and the communication port of the exposed service. |
| **spec.service.host** | **YES** | Specifies the service's external inbound communication address. |
| **spec.rules** | **YES** | Specifies array of rules. |
| **spec.rules.path** | **YES** | Specifies the path of the exposed service. |
| **spec.rules.methods** | **NO** | Specifies the list of HTTP request methods available for **spec.rules.path**. |
| **spec.rules.mutators** | **NO** | Specifies array of [Oathkeeper mutators](https://www.ory.sh/docs/oathkeeper/pipeline/mutator). |
| **spec.rules.accessStrategies** | **YES** | Specifies array of [Oathkeeper authenticators](https://www.ory.sh/docs/oathkeeper/pipeline/authn). |

## Additional information

When you fetch an existing APIRule CR, the system adds the **status** section which describes the status of the Virtual Service and the Rule created for this CR. This table lists the fields of the **status** section.

| Field   |  Description |
|:---|:---|
| **status.apiRuleStatus** | Status code describing the APIRule CR. |
| **status.virtualServiceStatus.code** | Status code describing the Virtual Service. |
| **status.virtualService.desc** | Current state of the Virtual Service. |
| **status.accessRuleStatus.code** | Status code describing the Oathkeeper Rule. |
| **status.accessRuleStatus.desc** | Current state of the Oathkeeper Rule. |

### Status codes

These are the status codes used to describe the Virtual Services and Rules:

| Code   |  Description |
|:---:|:---|
| **OK** | Resource created. |
| **SKIPPED** | Skipped creating a resource. |
| **ERROR** | Resource not created. |
