## Providers ##

The terraform modules specify in "version.tf" file what providers they need. Based on that informatioin the [entigo-infralib-agent](https://github.com/entigolabs/entigo-infralib-agent) will assemlbe the provider.tf file. Sometimes providers need extra configuration in that case a file with the same name as the provider is taken from this folder and included in the generated terraform code. 
