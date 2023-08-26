#!/bin/sh

# $1 == lambda's package
# $2 == target zip file

set -e

work=$(mktemp -d)

GOOS=linux CGO_ENABLED=0 GOARCH=arm64 go build -trimpath -o "$work"/bootstrap -tags lambda.norpc "$1"
touch -t 202111030000 "$work"/bootstrap
zip -Xj "$2" "$work"/bootstrap
rm -r "$work"