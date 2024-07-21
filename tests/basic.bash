#!/bin/bash



echo "Creating an instance of a server"

go install github.com/diarmuidmalanaphy/networktools@latest
go test .

