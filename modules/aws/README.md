## Terraform modules that are specific to AWS ##

__nuke.sh__  if runs locally then will first show what will be destroyed in entigo-infralib AWS account and then promts for confirmation. if runs in github actions then it will not promt and destroys all resources.
This helps

__aws-nuke-config.yml__ configuration of AWS Nuke - mostly needed to exclude some resources that won't be nuked every day in entigo-infralib AWS account.
