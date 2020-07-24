#!/bin/bash

UI_BUCKET=$(aws ssm get-parameter --output json --name pointypoints.uibucket | jq .Parameter.Value -r)

WORKSPACE=`(cd ../infrastructure && terraform workspace show)`

echo "Workspace: $WORKSPACE"
if [ $WORKSPACE != 'default' ]; then
  UI_BUCKET="$WORKSPACE-$UI_BUCKET"
fi

echo "Syncing dist/ to ${UI_BUCKET}"

# put everything in the bucket with a max age of 1 year
aws s3 sync ./dist "s3://$UI_BUCKET" --cache-control max-age=31536000 --delete --acl public-read
# then switch the max age on index.html to 60 seconds. Note, if stuff goes wrong with this or if something happens
# to hit cloudflare at the exact moment its cache expires its possible that cloudflare will cache it for a very long
# time, and we will need to invalidate the cache.
aws s3 cp "s3://$UI_BUCKET/index.html" "s3://$UI_BUCKET/index.html" --metadata-directive REPLACE --cache-control max-age=60 --content-type text/html --acl public-read