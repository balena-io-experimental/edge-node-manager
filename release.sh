#!/bin/bash

set -o errexit

if [ -z "$ACCOUNT" ] || [ -z "$REPO" ] || [ -z "$ACCESS_TOKEN" ]; then
	echo "Please set value for ACCOUNT, REPO and ACCESS_TOKEN!"
	exit 1
fi

# Create a release
rm -f request.json response.json
cat > request.json <<-EOF
{
    "tag_name": "$TRAVIS_TAG",
    "name": "$TRAVIS_TAG",
    "body": "Release of version $TRAVIS_TAG.\n$1"
}
EOF

curl --data "@request.json" --header "Content-Type:application/json" \
	"https://api.github.com/repos/$ACCOUNT/$REPO/releases?access_token=$ACCESS_TOKEN" \
	-o response.json
