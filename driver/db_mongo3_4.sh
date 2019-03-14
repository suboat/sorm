#!/bin/bash
PWD_NOW=$PWD
CON_NAME=orm-test-mongo
docker stop ${CON_NAME} > /dev/null 2>&1
docker rm ${CON_NAME} > /dev/null 2>&1
docker run -d \
  --restart=always \
  --name ${CON_NAME} \
  -p 27017:27017 \
  -e TZ=Asia/Shanghai \
  mongo:3.4
# waiting for init
sleep 5
#-e MONGO_INITDB_ROOT_USERNAME=business \
#-e MONGO_INITDB_ROOT_PASSWORD=business \
