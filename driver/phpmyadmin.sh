#!/bin/bash
docker stop phpadmin
docker rm phpadmin
docker run -d \
--name phpadmin \
--link orm-test-my:db \
-p 127.0.0.1:9000:80 \
phpmyadmin/phpmyadmin
