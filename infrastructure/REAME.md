# Pointypoints infrastructure

## Prerequisites

This terraform can be used to create the required infrastructure for pointypoints. Before executing terraform a domain name will need to be registered within Route53 (the app is not tied to the domain), and then the following parameters should be added to SSM:

* pointypoints.domainname - this should be the domain name registered (for example pointypoints.com)
* pointypoints.uibucket - this should be a name of an S3 bucket to host the UI out of (behind cloudfront of course). This bucket should not exist and will be created, it is paramiterized since bucket names must be globally unique.

## Executing

Ensure all of the lambda code has been built (execute `make` from the top level project directory) and ensure your aws cli env is setup to point to the desired account. Then execute `terraform apply`. Because the various TLS certs utilized for pointypoints are created via terraform they will take some time to verify, so some creates will fail as certificates will not be found. Wait for verification to complete and then execute `terraform apply` again and you should be in business.

## Workspaces

This project has been setup with workspace support, to spin up a workspace execute `terraform workspace new {whatever}` followed by `terraform apply`. Workspace creation is also subject to the certificate validation constraint so you will need to wait for certs to validate and then execute a second terraform apply.

## Gotchas

AWS stage deployment sucks. With V1 its possible to add variables to the deployment to force a deployment every time terraform runs, but alas this is not possible with the V2 api. So, if you add routes or anything execute:

`terraform taint aws_apigatewayv2_deployment.pointing`

and then apply to force a new deployment.