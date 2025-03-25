#!/bin/bash
set -e

pgqd -d /etc/pgqd.ini &

exec docker-entrypoint.sh "$@"
