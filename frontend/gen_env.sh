#!/bin/bash

DOMAIN=$(aws ssm get-parameter --output json --name pointypoints.domain | jq .Parameter.Value -r)

echo "VUE_APP_API_BASE_URL=https://api.${DOMAIN}" >> .env.local