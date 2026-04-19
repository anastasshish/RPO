#!/bin/sh
set -e

mkdir -p /app/data

/app/server &
nginx -g "daemon off;"
