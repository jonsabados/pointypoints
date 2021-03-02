# pointypoints

A serverless web based tool for task estimation. This code backs https://pointypoints.com, however it is not domain specific and could be used on any arbitrary domain registered within Route53. To deploy your very own pointypoints instance insure you have go and npm installed, then execute `make`, follow the instructions within the [infrastructure README](infrastructure/README.md) and then hop into the frontend directory and execute ./bucket_sync.sh

## Deploying code changes

### Backend

All backend code is deployed in the form of lambdas managed by terraform. To push changes execute `make` within the top level directory and then `terraform apply` within the infrastructure directory.

### Frontend

The UI is a Vue.js javascript application and may be run locally by executing `npm run serve` within the frontend directory. To deploy changes execute `make` from the top level directory and then `./bucket_sync.sh` from within the frontend directory.

## Executing tests

First have docker installed as it is used to run a local dynamo emulator. Then execute `make test` which will run unit tests for both the go code and frontend code.

## Outstanding issues
pointypoints was put together in a hurry as part of a time boxed learning exercise aimed mainly at playing with websockets + API gateway. As such it has some rough edges which should probably be addressed at some point. These include:
* Overuse of websockets. All communication with the backend happens via websockets, which actually makes life harder than it needs to be when it comes to dealing with requests that failed and whatnot - if things like writing to the datastore fail there is a good chance the user will face a spinner that never goes away - this is due to websockets not really being a request/response thing and rather just a way of sending messages back and forth without ties between messages. It would be better if all actions happend via more standard rest calls, and then websockets were only used to broadcast changes in state.
* Truly horrible persistence. Right now sessions are simply stored as giant blobs of json within Dynamo and require distributed locks to prevent concurrent updates. Some time and thought should actually be put into the persistence layer.
* Test coverage is lacking. Like if it were something done on the job I would be deeply ashamed level of lacking.
* ???