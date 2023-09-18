## Terraform modules that are specific to AWS ##

__nuke.sh__  if runs locally then will first show what will be destroyed in entigo-infralib AWS account and then promts for confirmation. if runs in github actions then it will not promt and destroys all resources.
This helps to keep costs under control and verify clean installation tests. It runs every day at 17:00 UTC.

__aws-nuke-config.yml__ configuration of AWS Nuke - mostly needed to exclude some resources that won't be nuked every day in entigo-infralib AWS account.


These modules can be used in the [entigo-infralib-agent](https://github.com/entigolabs/entigo-infralib-agent) steps of "__type: terraform__"

## Example code ##
```
steps:
  - name: network
    type: terraform
    workspace: test
    approve: minor
    modules:
      - name: hello
        source: aws/hello-world
        version: stable

```
