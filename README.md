# Custom-Controller
Just a small custom k8s controller

## Description
Builds a deployment from a custom resource and runs a pod containing the image specified, 
and, optionally, a redis cache.

## Getting Started

### Prerequisites
- go version v1.21.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### Deploy to cluster
**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=ghcr.io/shilohstuart6/custom-controller:latest
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin 
privileges or be logged in as admin.

**Create custom resources:**

Apply all samples from config/sample

```sh
kubectl apply -k config/samples/
```

or apply individually

```sh
kubectl apply -f config/samples/whatever_myappresource.yaml
```

### To Uninstall
**Delete the custom resources from the cluster:**

```sh
kubectl delete -k config/samples/
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

### Test
**Run the test suite:**
```sh
make test
```

## Alternate Installation
**Deploy the controller**

```sh
kubectl apply -f https://raw.githubusercontent.com/shilohstuart6/Custom-Controller/main/dist/install.yaml
```

**Deploy sample custom resource**

```sh
kubectl apply -f https://raw.githubusercontent.com/shilohstuart6/Custom-Controller/main/config/samples/whatever_myappresource.yaml
```
**Delete the controller**

```sh
kubectl delete -f https://raw.githubusercontent.com/shilohstuart6/Custom-Controller/main/dist/install.yaml
```

**Delete sample custom resource**

```sh
kubectl delete -f https://raw.githubusercontent.com/shilohstuart6/Custom-Controller/main/config/samples/whatever_myappresource.yaml
```

## Build Commands
**Build and push image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=ghcr.io/shilohstuart6/custom-controller:latest
```

**Build installer for the image:**
```sh
make build-installer IMG=ghcr.io/shilohstuart6/custom-controller:latest
```

**NOTE:** This image ought to be published in the personal registry you specified. 
And it is required to have access to pull the image from the working environment. 
Make sure you have the proper permission to the registry if the above commands don’t work.

## License

Copyright 2024 shiliohstuart6.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

