# Pointypoints infrastructure

This terraform can be used to create the required infrastructure for pointypoints. You should execute

* pointypoints.domainname
* pointypoints.uibucket

## Workspaces

This project has been setup with workspace support, to spin up a workspace execute `terraform workspace new {whatever}` followed by `terraform apply` - the apply will fail as it creates an ACM cert that will take time to validate. Once the cert has been validated run `terraform apply` again.