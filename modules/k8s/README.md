## Helm charts that we use ##

These modules can be used in the [entigo-infralib-agent](https://github.com/entigolabs/entigo-infralib-agent) steps of "__type: argocd-apps__".


## Example code ##
```
steps:
  - name: apps
    type: argocd-apps
    modules:
      - name: hello-world
        source: hello-world
        version: stable

```
