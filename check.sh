#!/bin/bash
BEGIN_DIR=$PWD
for dirName in "." "log" "types" "driver" "songo" "driver/pg" "driver/mysql" "driver/sqlite" "driver/mongo"; do
  echo checking... ${dirName}
  cd ${BEGIN_DIR}/${dirName}
  # code fmt
  go fmt .
  # buildin tools
  go tool vet .
  # get https://github.com/golang/lint
  golint .
  # get https://github.com/dominikh/go-tools/releases
  staticcheck .
done
