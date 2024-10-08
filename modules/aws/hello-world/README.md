## Dummy module for testing ##
This module is created for terraform pipeline testing. It is meant to plan and apply fast (no providers and no resources created).

It outputs the text "Hello, ${var.prefix}!"
It has no input variables




### Example code ###

```
    modules:
      - name: hello
        source: aws/hello-world

```
