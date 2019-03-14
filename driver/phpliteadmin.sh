#!/bin/bash
mkdir -p $PWD/data_sqlite
docker stop phpadmin
docker rm phpadmin
docker run -d \
--name phpadmin \
-v $PWD/data_sqlite:/db \
-p 127.0.0.1:9000:2015 \
acttaiwan/phpliteadmin

# next:
# 1. vist http://127.0.0.1:9000/phpliteadmin.php
# 2. password: admin
