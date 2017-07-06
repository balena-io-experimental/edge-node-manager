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

# Check the id
if [ $id = "null" ]; then
    # Release does not already exist so we create a new one
    echo "Creating a new $TRAVIS_TAG release"

    json="{
        \"tag_name\": \"$TRAVIS_TAG\",
        \"name\": \"$TRAVIS_TAG\",
        \"body\": \"[Changelog](https://github.com/resin-io/edge-node-manager/blob/master/CHANGELOG.md)\n$release\"
    }"

    curl --data "$json" --header "Content-Type:application/json" \
	    "https://api.github.com/repos/$ACCOUNT/$REPO/releases?access_token=$ACCESS_TOKEN"
else
    # Release already exists so we append to the existing body
    echo "Appending to existing $TRAVIS_TAG release"

    json="{
	\"body\": \"$body\n$release\"
    }"

    curl --data "$json" --header "Content-Type:application/json" \
	-X PATCH "https://api.github.com/repos/$ACCOUNT/$REPO/releases/$id?access_token=$ACCESS_TOKEN"
fi
