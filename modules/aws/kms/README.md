## Oppinionated module for kms creation ##


Oppinionated version of this https://registry.terraform.io/modules/terraform-aws-modules/kms/aws/latest

The other modules will automatically detect the KMS keys and encrypt their data.

The module will create 3 keys: telemetry, configuration and data.

### Example code ###
```
    modules:
      - name: kms
        source: aws/kms
```

