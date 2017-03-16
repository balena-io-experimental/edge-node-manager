#!/bin/bash

set -o errexit

if [ -z "$ACCOUNT" ] || [ -z "$REPO" ] || [ -z "$ACCESS_TOKEN" ]; then
	echo "Please set value for ACCOUNT, REPO and ACCESS_TOKEN!"
	exit 1
fi

json="{
    \"tag_name\": \"$TRAVIS_TAG\",
    \"name\": \"$TRAVIS_TAG\",
    \"body\": \"Release of $TRAVIS_TAG.\n$1\"
}"

curl --data "$json" --header "Content-Type:application/json" \
    "https://api.github.com/repos/$ACCOUNT/$REPO/releases?access_token=$ACCESS_TOKEN"
