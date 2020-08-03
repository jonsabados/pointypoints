#/bin/sh

if docker container ls | grep dynamo > /dev/null
then
  # redis is already running, we can just execute the tests and call it a day
  go test -v ./... --race --cover
else
  # need to start redis, execute tests & retain exit code, shutdown redis and make sure the tests exit code is returned
  docker run -d --name dynamo -p 8000:8000 amazon/dynamodb-local 2>&1 >/dev/null
  go test -v ./... --race --cover
  ret=$?
  docker stop dynamo 2>&1 >/dev/null
  docker rm dynamo 2>&1 >/dev/null
  exit ${ret}
fi