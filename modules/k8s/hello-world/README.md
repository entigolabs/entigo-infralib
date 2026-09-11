## Dummy module for testing ##
This module is created for ArgoCD/Helm pipeline testing. It launched very basic Kubernetes application.
No additional values need to be specified.
<!-- trigger: testing arc-runner-set migration for k8s PR workflow -->

### Example code ###

```
    modules:
      - name: hello-world
        source: hello-world

```
