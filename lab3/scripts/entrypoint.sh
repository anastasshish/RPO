#!/bin/sh
set -e
/app/server &
nginx -g 'daemon off;'
