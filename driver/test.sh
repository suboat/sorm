#!/bin/bash
#./db_postgres11_2.sh
#export GO111MODULE=off
rm -f test.log.*
go test -failfast -cover -v
go test -failfast -bench=. -run=None -benchtime=1s
#go test -failfast -cover -v -run=Objects
