#!/usr/bin/env bash

printf "Build Started: package-finder binary \n"

go build -o bin/package-finder

printf "Build succeeded: bin/package-finder \n"