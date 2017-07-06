#!/bin/bash

set -o errexit

# Check all the required variables are defined
if [ -z "$ACCOUNT" ] || [ -z "$REPO" ] || [ -z "$ACCESS_TOKEN" ] || [ -z "$TRAVIS_TAG" ]; then
	echo "Please set value for ACCOUNT, REPO, ACCESS_TOKEN and TRAVIS_TAG"
	exit 1
fi

# Request the release as it may already exist
release=$(curl https://api.github.com/repos/$ACCOUNT/$REPO/releases/tags/$TRAVIS_TAG)

# Extract the id and the body
id=$(echo $release | jq .id)
body=$(echo $release | jq .body)

# Remove quotes from around the body
body="${body%\"}"
body="${body#\"}"

# Get the release string from travis (AWS download link)
release=$1

# Debugging print
echo ""
echo "START DEBUG"
echo "id: $id"
echo "body: $body"
echo "release: $release"
echo "END DEBUG"
echo ""

# Attempt to create a new release instead of checking whether the release already exists
# This avoids situations where the release is created by another job mid way through this script
echo "Attempting to create a new $TRAVIS_TAG release"
json="{
    \"tag_name\": \"$TRAVIS_TAG\",
    \"name\": \"$TRAVIS_TAG\",
    \"body\": \"Release of $TRAVIS_TAG: [changelog](https://github.com/resin-io/edge-node-manager/blob/master/CHANGELOG.md)\n$release\"
}"

resp=$(curl -i --data "$json" --header "Content-Type:application/json" \
	"https://api.github.com/repos/$ACCOUNT/$REPO/releases?access_token=$ACCESS_TOKEN" | \
	head -n 1 | cut -d$' ' -f2)

# Handle the response
if [ $resp = "201" ]; then
    echo "Success"
elif [ $resp = "422" ]; then
    echo "Release already exists, appending instead"

    json="{
	\"body\": \"$body\n$release\"
    }"

    curl --data "$json" --header "Content-Type:application/json" \
	-X PATCH "https://api.github.com/repos/$ACCOUNT/$REPO/releases/$id?access_token=$ACCESS_TOKEN"
fi
