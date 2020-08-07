# pointypoints

A serverless web based tool for task estimation. This code backs https://pointypoints.com, however it is not domain specific and could be used on any arbitrary domain registered within Route53. To deploy your very own pointypoints instance insure you have go and npm installed, then execute `make`, follow the instructions within the [infrastructure README](infrastructure/README.md) and then hop into the frontend directory and execute ./bucket_sync.sh

## Deploying code changes

### Backend

All backend code is deployed in the form of lambdas managed by terraform. To push changes execute `make` within the top level directory and then `terraform apply` within the infrastructure directory.

### Frontend

The UI is a Vue.js javascript application and may be run locally by executing `npm run serve` within the frontend directory. To deploy changes execute `make` from the top level directory and then `./bucket_sync.sh` from within the frontend directory.

## Executing tests

First have docker installed as it is used to run a local dynamo emulator. Then execute `make test` which will run unit tests for both the go code and frontend code.