#! /bin/bash
echo "Builing main.go to main"
rm -f main
go build *.go
#go build -ldflags "-s -w" *.go
echo "Running"
./main
