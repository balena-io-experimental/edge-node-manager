#!/bin/bash

set -o errexit

if [ -z "$ACCOUNT" ] || [ -z "$REPO" ] || [ -z "$ACCESS_TOKEN" ] || [ -z "$TRAVIS_TAG" ]; then
	echo "Please set value for ACCOUNT, REPO, ACCESS_TOKEN and TRAVIS_TAG"
	exit 1
fi

echo "Attempting to create a new $TRAVIS_TAG release"
json="{
	\"tag_name\": \"$TRAVIS_TAG\",
	\"name\": \"$TRAVIS_TAG\",
	\"body\": \"Release of $TRAVIS_TAG: [changelog](https://github.com/resin-io/edge-node-manager/blob/master/CHANGELOG.md)\n$1\"
}"

resp=$(curl -i --data "$json" --header "Content-Type:application/json" \
	"https://api.github.com/repos/$ACCOUNT/$REPO/releases?access_token=$ACCESS_TOKEN" | \
	head -n 1 | cut -d$' ' -f2)

if [ $resp = "201" ]; then
	echo "Success"
elif [ $resp = "422" ]; then
	echo "Release already exists, appending instead"

	release=$(curl https://api.github.com/repos/$ACCOUNT/$REPO/releases/tags/$TRAVIS_TAG)
	id=$(echo $release | jq .id)
	body=$(echo $release | jq .body)
	body="${body%\"}"
	body="${body#\"}"

	json="{
		\"body\": \"$body\n$1\"
	}"

	resp=$(curl --data "$json" --header "Content-Type:application/json" \
		-X PATCH "https://api.github.com/repos/$ACCOUNT/$REPO/releases/$id?access_token=$ACCESS_TOKEN" | \
		head -n 1 | cut -d$' ' -f2)

	if [ $resp = "200" ]; then
		exit 0
	else
		exit 1
	fi
fi
