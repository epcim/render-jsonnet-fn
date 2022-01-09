# render-jsonnet

An KRM Function to render jsonnet configurations, manifests...

- [KRM Fn specification](https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md)
- [go-getter](https://github.com/hashicorp/go-getter) is used to fetch sources
- jsonnet

```sh
# build
docker build -t render-jsonnet . 

# usage
kustomize build --enable-alpha-plugins --network ./example 

# dev
kustomize build --enable-alpha-plugins --network example --mount "type=bind,rw=true,src=$PWD/tmp,dst=/tmp"
```


## Dev

- https://github.com/google/go-jsonnet
- https://github.com/jsonnet-bundler/jsonnet-bundler
