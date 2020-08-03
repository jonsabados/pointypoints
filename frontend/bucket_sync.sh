#!/bin/bash

UI_BUCKET=$(aws ssm get-parameter --output json --name pointypoints.uibucket | jq .Parameter.Value -r)

WORKSPACE=`(cd ../infrastructure && terraform workspace show)`

if [ $WORKSPACE != 'default' ]; then
  UI_BUCKET="$WORKSPACE-$UI_BUCKET"
fi

echo "Syncing dist/ to ${UI_BUCKET}"

# When syncing the idea is put all the new stuff in, then put the new index page in that references the new stuff
# and then finally purge the old stuff. This way there isn't any funkyness with getting an index page that references
# assets that haven't been uploaded, or CloudFront caching the index for long periods of time if the max age on it
# hasn't been set
aws s3 sync ./dist "s3://$UI_BUCKET" --exclude index.html  --cache-control max-age=31536000 --acl public-read
aws s3 cp ./dist/index.html "s3://$UI_BUCKET/index.html" --metadata-directive REPLACE --cache-control max-age=60 --content-type text/html --acl public-read
aws s3 sync ./dist "s3://$UI_BUCKET" --exclude index.html --cache-control max-age=31536000 --delete --acl public-read