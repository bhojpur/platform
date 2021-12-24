#!/usr/bin/env bash

DOMAIN=app.bhojpur.net
EMAIL=info@bhojpur.net

# shellcheck disable=SC2035
docker run -it --rm -v "$(pwd)"/letsencrypt:/letsencrypt --user "$(id -u)":"$(id -g)" certbot/certbot certonly \
    --config-dir /letsencrypt/config \
    --work-dir /letsencrypt/work \
    --logs-dir /letsencrypt/logs \
    --manual \
    --preferred-challenges=dns \
    --email $EMAIL \
    --agree-tos \
    -d $DOMAIN \
    -d *.$DOMAIN \
    -d *.gitlab.$DOMAIN \
    -d *.bhojpur.$DOMAIN \
    -d *.ws.bhojpur.$DOMAIN


find letsencrypt/config/live -name "*.pem" -exec cp {} ./ \;

openssl dhparam -out dhparams.pem 2048
