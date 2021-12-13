#!/bin/bash
PWD_NOW=$PWD
CON_NAME=orm-test-pg
docker stop ${CON_NAME} > /dev/null 2>&1
docker rm ${CON_NAME} > /dev/null 2>&1
docker run -d \
  --restart=always \
  --name ${CON_NAME} \
  -p 65432:5432 \
  -e POSTGRES_PASSWORD=postgres \
  -e TZ=Asia/Shanghai \
  -v ${PWD_NOW}/init_postgres9_6.sh:/docker-entrypoint-initdb.d/init-user-db.sh \
  postgres:12-alpine
# waiting for init
sleep 5
