## Opinionated module for config-rules creation

`./yaml-to-tf.sh` will convert AWS Config rules from CloudFormation to Terraform and write them to rules.tf

### Example code

```
    modules:
      - name: config-rules
        source: aws/config-rules

```
