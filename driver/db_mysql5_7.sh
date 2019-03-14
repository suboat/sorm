#!/bin/bash
PWD_NOW=$PWD
CON_NAME=orm-test-my
docker stop ${CON_NAME} > /dev/null 2>&1
docker rm ${CON_NAME} > /dev/null 2>&1
docker run -d \
  --restart=always \
  --name ${CON_NAME} \
  -p 33306:3306 \
  -e MYSQL_ROOT_PASSWORD=mysql \
  -e MYSQL_DATABASE=business \
  -e MYSQL_USER=business \
  -e MYSQL_PASSWORD=business \
  -e TZ=Asia/Shanghai \
  mysql:5.7
# waiting for init
sleep 5
