# render-jsonnet

An KRM Function to render go templated manifests. 

- [KRM Fn specification](https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md)
- [go-getter](https://github.com/hashicorp/go-getter) is used to fetch sources
- jsonnet + sprig rendering

TODO:
- render engine [gomplate](https://gomplate.ca/) is used to render templates

```sh
# build
docker build -t render-jsonnet . 

# usage
kustomize build --enable-alpha-plugins --network ./example 

# dev
kustomize build --enable-alpha-plugins --network example --mount "type=bind,rw=true,src=$PWD/tmp,dst=/tmp"
```


## Engines

- Jsonnet
- ~Gomplate~
- ~Helm~

### Gomplate

Not implemented

### Helm

Not implemented (send MR if you want it)
