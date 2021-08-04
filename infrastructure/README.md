# Pointypoints infrastructure

## Prerequisites

This terraform can be used to create the required infrastructure for pointypoints. Before executing terraform a domain name will need to be registered within Route53 (the app is not tied to the domain), and then the following parameters should be added to SSM:

* pointypoints.domainname - this should be the domain name registered (for example pointypoints.com)
* pointypoints.uibucket - this should be a name of an S3 bucket to host the UI out of (behind cloudfront of course). This bucket should not exist and will be created, it is paramiterized since bucket names must be globally unique.
* pointypoints.google.verification - this should be the site verification txt record for google domain ownership verification. Any value can be placed in this if that is not a concern (note, it will show up in the txt record for the domain)
* pointypoints.google.clientId - this should be a client id for google web sign in

Next you will need to create a bucket to store the terraform state. There are no constraints on the bucket name, but the bucket should be private (it will function if its public but exposing it to the world is a terrible idea).

## Terraform config

The provider configuration for AWS is configured to use default credentials, so you will need to get your environment setup so all the various aws cli commands point to whatever account you are deploying to. Running `aws configure` is one way to accomplish this. 

Once you can do things like execute `aws s3 ls` and see the bucket you have created for state you are good to run `terraform init` within this directory. You will be prompted for the bucket to store state in - poke in the name of the bucket you created. If you are collaborating with any other individuals within the same account just make sure you all use a common state bucket or all sorts of oddness will abound.

## Executing

Ensure all of the lambda code has been built (execute `make` from the top level project directory), then execute `terraform apply`.

It may take some time for DNS to propagate, but once that happens you should be good to point a browser at https://{yourdomain} or https://www.{yourdomain}

## Workspaces

This project has been setup with workspace support, to spin up a workspace execute `terraform workspace new {whatever}` followed by `terraform apply`. When creating a workspace the application will be deployed to https://{workspace}.{yourdomain} and https://www-{workspace}.{yourdomain}

#### DANGER!!!

The workspace is referenced within the front ends .env.local file, so if you switch workspaces run a `make clean build` to
update that
