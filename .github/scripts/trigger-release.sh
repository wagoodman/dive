#!/usr/bin/env bash
set -eu

bold=$(tput bold)
normal=$(tput sgr0)

if ! [ -x "$(command -v gh)" ]; then
    echo "The GitHub CLI could not be found. To continue follow the instructions at https://github.com/cli/cli#installation"
    exit 1
fi

gh auth status

# we need all of the git state to determine the next version. Since tagging is done by
# the release pipeline it is possible to not have all of the tags from previous releases.
git fetch --tags

# populates the CHANGELOG.md and VERSION files
echo "${bold}Generating changelog...${normal}"
make changelog 2> /dev/null

NEXT_VERSION=$(cat VERSION)

if [[ "$NEXT_VERSION" == "" ||  "${NEXT_VERSION}" == "(Unreleased)" ]]; then
    echo "Could not determine the next version to release. Exiting..."
    exit 1
fi

while true; do
    read -p "${bold}Do you want to trigger a release for version '${NEXT_VERSION}'?${normal} [y/n] " yn
    case $yn in
        [Yy]* ) echo; break;;
        [Nn]* ) echo; echo "Cancelling release..."; exit;;
        * ) echo "Please answer yes or no.";;
    esac
done

echo "${bold}Kicking off release for ${NEXT_VERSION}${normal}..."
echo
gh workflow run release.yaml -f version=${NEXT_VERSION}

echo
echo "${bold}Waiting for release to start...${normal}"
sleep 10

set +e

echo "${bold}Head to the release workflow to monitor the release:${normal} $(gh run list --workflow=release.yaml --limit=1 --json url --jq '.[].url')"
id=$(gh run list --workflow=release.yaml --limit=1 --json databaseId --jq '.[].databaseId')
gh run watch $id --exit-status || (echo ; echo "${bold}Logs of failed step:${normal}" && GH_PAGER="" gh run view $id --log-failed)
