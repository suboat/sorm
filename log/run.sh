#!/bin/bash
rm -rf ./log
rm -f test.log*
rm -f all.log
go test -v
ls -al
ls -al ./log
