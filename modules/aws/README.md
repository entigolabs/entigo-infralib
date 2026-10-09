## OpenTofu modules that are specific to AWS ##

The daily nuke of the entigo-infralib AWS account is configured in [nuke/aws.yaml](../../nuke/README.md).


These modules can be used in the [entigo-infralib-agent](https://github.com/entigolabs/entigo-infralib-agent) steps of "__type: terraform__"

## Example code ##
```
steps:
  - name: network
    type: terraform
    workspace: test
    modules:
      - name: hello
        source: aws/hello-world

```
