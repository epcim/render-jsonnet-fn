# render-jsonnet

An KRM Function to render jsonnet configurations, manifests...
An generator to be used with Kubectl, Kustomize or Kpt...

## Usage: Shell implementation

An prototype.

```
kustomize build --enable-alpha-plugins --enable-exec ./example-exec

# with other flags
kustomize build --enable-alpha-plugins --network --enable-exec --load-restrictor LoadRestrictionsNone ./example-exec
```


## Usage: Go implementation

TBD

## Function

[KRM Fn specification](https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md)

See upstream/other function examples:
- https://github.com/GoogleContainerTools/kpt-functions-catalog/blob/master/functions

My other functions:
- https://github.com/epcim/render-gotpl-fn


## Setup to run examples

```
brew install jsonnet-bundler yq jq kustomize kubectl

jb install

# dependencies
# jb install github.com/kubernetes-monitoring/kubernetes-mixin
# jb install github.com/jsonnet-libs/k8s-libsonnet/1.24@main

```

## Dev

- https://github.com/google/go-jsonnet
- https://github.com/jsonnet-bundler/jsonnet-bundler


```
go install github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb@latest
brew install jsonnet

cd example-exec
git clone https://github.com/kubernetes-monitoring/kubernetes-mixin

cd  kubernetes-mixin
jb install

cd ../..
cd example-exec
../fnRenderJsonnet kubnernets-mixin/lib/alerts.json
```
