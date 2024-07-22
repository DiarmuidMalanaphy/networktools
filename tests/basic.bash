#!/bin/bash

echo "Grabbing latest version from the git"
go get -u github.com/diarmuidmalanaphy/networktools@latest

echo "Running tests"

go test . -v

