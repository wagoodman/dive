#!/usr/bin/env bash

red=$(tput setaf 1)
bold=$(tput bold)
normal=$(tput sgr0)

# assert we are running in CI (or die!)
if [[ -z "$CI" ]]; then
    echo "${bold}${red}This step should ONLY be run in CI. Exiting...${normal}"
    exit 1
fi
